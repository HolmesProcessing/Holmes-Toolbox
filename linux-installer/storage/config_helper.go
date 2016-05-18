package main

import (
    "bufio"
    "fmt"
    "os/exec"
    "sync"
    "strings"
    "encoding/json"
    "flag"
    "os"
    "strconv"
    "errors"
    "time"
    
    "github.com/gocql/gocql"
    "github.com/HolmesProcessing/Holmes-Storage/objStorerGeneric"
    "github.com/HolmesProcessing/Holmes-Storage/storerGeneric"
)

/*
* struct as presented in Holmes-Storage/main.go
*/
type config struct {
    Storage     string
    Database    []*storerGeneric.DBConnector
    ObjStorage  string
    ObjDatabase []*objStorerGeneric.ObjDBConnector
    
    LogFile     string
    LogLevel    string

    AMQP          string
    Queue         string
    RoutingKey    string
    PrefetchCount int

    HTTP         string
    ExtendedMime bool
}

/*
* Helper to gather user inputs
*/
var (
    msg_logfile         string = "Path for a logfile, empty for no log"
    msg_loglevel        string = "Loglevel"
    msg_rabbitip        string = "RabbitMQ IP-Address"
    msg_rabbitport      string = "RabbitMQ Port"
    msg_rabbituser      string = "RabbitMQ Username"
    msg_rabbitpassword  string = "RabbitMQ Password"
    msg_queue           string = "RabbitMQ queue to fetch results from"
    msg_routingkey      string = "RabbitMQ routing key"
    msg_prefetchcount   string = "RabbitMQ item prefetch count"
    msg_storage_ip      string = "Holmes-Storage IP-Address"
    msg_storage_port    string = "Holmes-Storage Port"
    msg_extendedmime    string = "Should extended mime type be parsed"
    msg_db_ip           string = "IP"
    msg_db_port         string = "Port"
    msg_db_user         string = "Username"
    msg_db_password     string = "Password (stored in plaintext!)"
    msg_db_database     string = "Database"
    msg_s3_ip           string = "IP"
    msg_s3_port         string = "Port"
    msg_s3_region       string = "Region"
    msg_s3_key          string = "Key"
    msg_s3_secret       string = "Secret"
    msg_s3_bucket       string = "Bucket"
    msg_s3_disablessl   string = "Disable SSL"
    stdin *bufio.Reader = nil
)
func get_text_generic(msg string, default_result string, options string) string {
    // keep only one global reference to stdin
    if stdin == nil {
        stdin = bufio.NewReader(os.Stdin)
    }
    // if there is a default text or answer options supplied, show them
    var text string
    if default_result != "" {
        text = fmt.Sprintf("default %s", default_result)
    }
    if options != "" {
        if text != "" {
            text = fmt.Sprintf("%s, %s", options, text)
        } else {
            text = fmt.Sprintf("%s", options)
        }
    }
    if text != "" {
        text = fmt.Sprintf(" (%s)", text)
    }
    // print user query and read input, but trim newlines and whitespaces
    fmt.Printf("  - %s%s: ", msg, text)
    result, _ := (stdin).ReadString('\n')
    result = strings.TrimSpace(result)
    // return default if user input is empty
    if result == "" {
        return default_result
    }
    return result
}
func get_text(msg string, default_text string) string {
    return get_text_generic(msg, default_text, "")
}
func get_positive_integer(msg string, default_text string) int {
    for {
        text := get_text(msg, default_text)
        result, err := strconv.ParseInt(text,10,32)
        if err == nil && result >= 0 {
            return int(result)
        } else {
            fmt.Println("    Invalid number entered, must be positive and Integer '", err, "'")
        }
    }
}
func get_boolean(msg string, default_value bool) bool {
    text := get_text_generic(msg, fmt.Sprintf("%t", default_value), "true/false")
    if text == "true" {
        return true
    } else {
        return false
    }
}

func get_dbconnector(username, password, database, defaultPort string) *storerGeneric.DBConnector {
    // get only IP and Port, since User,Password and Database must match for Cassandra,
    // MongoDB setup calls User+Password request additionally
    db := &storerGeneric.DBConnector{}
    db.IP = get_text(msg_db_ip, "127.0.0.1")
    db.Port = get_positive_integer(msg_db_port, defaultPort)
    db.User = username
    db.Password = password
    db.Database = database
    return db
}
func get_objdbconnector() *objStorerGeneric.ObjDBConnector {
    // Each S3 box can be set up individually, as such ask for all credentials
    // over and over again
    objdb := &objStorerGeneric.ObjDBConnector{}
    objdb.IP = get_text(msg_s3_ip, "")
    objdb.Port = get_positive_integer(msg_s3_port, "27017")
    objdb.Region = get_text(msg_s3_region, "us-east-1")
    objdb.Key = get_text(msg_s3_key, "")
    objdb.Secret = get_text(msg_s3_secret, "")
    objdb.Bucket = get_text(msg_s3_bucket, "holmes_totem")
    objdb.DisableSSL = get_boolean(msg_s3_disablessl, true)
    return objdb
}

func get_database(defaultDatabase string) string {
    return get_text(msg_db_database, defaultDatabase)
}
func get_username_password(defaultUser, defaultPassword string) (user, password string) {
    user = get_text(msg_db_user, defaultUser)
    password = get_text(msg_db_password, defaultPassword)
    return
}
func get_key_secret(defaultKey, defaultSecret string) (key, secret string) {
    key = get_text(msg_s3_key, "")
    secret = get_text(msg_s3_secret, "")
    return
}

func get_cassandra_boxes(limit int) []*storerGeneric.DBConnector {
    var dbs []*storerGeneric.DBConnector
    i := 1
    // username, password and database (keyspace) must match for all boxes
    fmt.Println("> Please supply a username, password and database to use with your Cassandra setup:")
    database := get_database("holmes_totem")
    username,password := get_username_password("cassandra","cassandra")
    // grab dbconnector for all boxes and apply the username/password/dbname on them
    for {
        fmt.Printf("> Add configuration for Cassandra box #%d:\n", i)
        dbs = append(dbs, get_dbconnector(username, password, database, "9042"))
        if (i == limit) || (! get_boolean("\n  Add another Cassandra box?", false)) {
            break
        }
        i += 1
    }
    return dbs
}
func get_mongodb_boxes(limit int) []*storerGeneric.DBConnector {
    var dbs []*storerGeneric.DBConnector
    i:= 1
    // database (keyspace) must match for all boxes (username + password can vary)
    fmt.Println("> Please supply a database to use with your MongoDB setup:")
    database := get_database("holmes_totem")
    // grab dbconnector, username and password for all boxes and apply dbname on them
    for {
        fmt.Printf("> Add configuration for MongoDB box #%d:\n", i)
        username, password := get_username_password("admin", "admin")
        dbs = append(dbs, get_dbconnector(username, password, database, "27017"))
        if (i == limit) || (! get_boolean("\n  Add another MongoDB box?", false)) {
            break
        }
        i += 1
    }
    return dbs
}

func get_s3_boxes(limit int) []*objStorerGeneric.ObjDBConnector {
    var dbs []*objStorerGeneric.ObjDBConnector
    i := 1
    for {
        fmt.Printf("> Add configuration for S3 storage #%d:\n", i)
        dbs = append(dbs, get_objdbconnector())
        if (i == limit) || (! get_boolean("\n  Add another S3 storage?", false)) {
            break
        }
        i += 1
    }
    return dbs
}

/*
* Helper to execute commands
*/
func execute(cmd string, wg *sync.WaitGroup) []byte {
    fmt.Println("executing:",cmd)
    // splitting head => g++ parts => rest of the command
    parts := strings.Fields(cmd)
    head := parts[0]
    parts = parts[1:len(parts)]
    // now use variadic arguments to get all params passed through to the command
    out, err := exec.Command(head,parts...).Output()
    if err != nil {
        fmt.Printf("%s", err)
    }
    fmt.Printf("%s", out)
    if wg != nil {
        wg.Done() // Need to signal to waitgroup that this goroutine is done
    }
    return out
}

/*
* Set up a local cassandra box (create keyspace)
* More lines of code, but Less error prone than handing it to execute and
* letting cqlsh set up the keyspace
*/
func setup_cassandra_keyspace(dbs []*storerGeneric.DBConnector) (error) {
    if len(dbs) < 1 {
        return errors.New("Supply at least one node to connect to!")
    }
    connStrings := make([]string, len(dbs))
    for i, elem := range dbs {
        connStrings[i] = fmt.Sprintf("%s:%d", elem.IP, elem.Port)
    }
    var err error
    cluster := gocql.NewCluster(connStrings...)
    cluster.Authenticator = gocql.PasswordAuthenticator{
        Username: dbs[0].User,
        Password: dbs[0].Password,
    }
    cluster.ProtoVersion = 4
    // It is important to connect without a keyspace selected, otherwise creating
    // a new session will error out
    cluster.Keyspace = ""
    cluster.Consistency = gocql.Quorum
    // if gocql complains about timeouts and the cassandra log file shows unusual
    // long operations, increase the clusters timeout value
    cluster.Timeout = 30 * time.Second
    DB, err := cluster.CreateSession()
    if err != nil {
        return err
    }
    // might not be important, but make sure the connection is terminated properly
    defer DB.Close()
    // create keyspace
    fmt.Printf("> Creating keyspace '%s' on all supplied Cassandra nodes\n", dbs[0].Database)
    cqlquery := fmt.Sprintf(`
        CREATE KEYSPACE IF NOT EXISTS %s WITH REPLICATION = {
            'class' : 'SimpleStrategy',
            'replication_factor' : 1
        };
        `, dbs[0].Database)
    if err := DB.Query(cqlquery).Exec(); err != nil {
        return err
    }
    return nil
}

/*
* Write the generated config to a file
*/
func write_config(configObj config) {
    cfile, err := os.OpenFile(
        "./config.json",
        os.O_CREATE | os.O_RDWR | os.O_TRUNC,
        0644)
    if err != nil {
        panic("Error opening ./config.json, please check filesystem permissions and disk space availability")
    }
    defer cfile.Close()
    b, _ := json.MarshalIndent(configObj,"","    ")
    cfile.Write(b)
    cfile.Write([]byte("\n"))
}

/*
* Create a valid config file and even set up a local cassandra if necessary
*/
func main() {
    var (
        configType      string
        configObj       config
    )
    // read command line arguments
    flag.StringVar(&configType, "config", "cluster", "The type of the config to create.")
    flag.Parse()
    
    // Gather database/storage settings
    switch configType {
        case "local":
            configObj.Storage = "cassandra"
            configObj.ObjStorage = "local-fs"
            configObj.Database = get_cassandra_boxes(1)
            err := setup_cassandra_keyspace(configObj.Database)
            if err != nil {
                panic(err)
            }
            
        case "local-objstorage":
            configObj.Storage = "cassandra"
            configObj.ObjStorage = "local-fs"
            configObj.Database = get_cassandra_boxes(-1)  // unlimited boxes
            err := setup_cassandra_keyspace(configObj.Database)
            if err != nil {
                panic(err)
            }
            
        case "local-cassandra":
            configObj.Storage = "cassandra"
            configObj.ObjStorage = "S3"
            configObj.Database = get_cassandra_boxes(1)
            configObj.ObjDatabase = get_s3_boxes(1)  // current limit seems to be 1 (see objStorerS3.go)
            err := setup_cassandra_keyspace(configObj.Database)
            if err != nil {
                panic(err)
            }
            
        case "cluster":
            configObj.Storage = "cassandra" 
            configObj.ObjStorage = "S3"
            configObj.Database = get_cassandra_boxes(-1)  // unlimited boxes
            configObj.ObjDatabase = get_s3_boxes(1)  // current limit seems to be 1 (see objStorerS3.go)
            err := setup_cassandra_keyspace(configObj.Database)
            if err != nil {
                panic(err)
            }
            
        case "local-mongodb":
            configObj.Storage = "mongodb"
            configObj.Database = get_mongodb_boxes(1)
            
        case "cluster-mongodb":
            configObj.Storage = "mongodb"
            configObj.Database = get_mongodb_boxes(-1)  // unlimited boxes
            
        default:
            fmt.Println("Invalid option for --config= supplied.")
            panic("Invalid option for --config= supplied.")
    }
    
    // gather various other information
    fmt.Println("> Other settings:")
    configObj.LogFile       = get_text(msg_logfile, "")
    configObj.LogLevel      = get_text_generic(msg_loglevel, "debug", "error/warn/info/debug")
    
    rabbitIP               := get_text(msg_rabbitip, "127.0.0.1")
    rabbitPort             := get_text(msg_rabbitport, "5672")
    rabbitAuth             := ""
    rabbitUser             := get_text(msg_rabbituser, "guest")
    if rabbitUser != "" {
        rabbitPassword     := get_text(msg_rabbitpassword, "guest")
        if rabbitPassword != "" {
            rabbitAuth = fmt.Sprintf("%s:%s@",rabbitUser,rabbitPassword)
        } else {
            rabbitAuth = fmt.Sprintf("%s@",rabbitUser)
        }
    }
    configObj.AMQP          = fmt.Sprintf("amqp://%s%s:%s",rabbitAuth,rabbitIP,rabbitPort)
    configObj.Queue         = get_text(msg_queue, "totem_output")
    configObj.RoutingKey    = get_text(msg_routingkey, "*.result.static.totem")
    configObj.PrefetchCount = get_positive_integer(msg_prefetchcount, "10")
    
    storageIP              := get_text(msg_storage_ip, "127.0.0.1")
    storageHTTP            := storageIP
    storagePort            := get_text(msg_storage_port, "8016")
    if storagePort != "" {
        storageHTTP = fmt.Sprintf("%s:%s", storageIP, storagePort)
    }
    configObj.HTTP          = storageHTTP
    configObj.ExtendedMime  = get_boolean(msg_extendedmime, true)
    
    // write the resulting config to file
    fmt.Println("> Writing config file to ./config.json")
    write_config(configObj)
    
    x := []string{"/bin/cat ./config.json"}
    execute(x[0], nil)
}
