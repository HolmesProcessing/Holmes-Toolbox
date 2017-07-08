import pika
import json, os

#TODO: add cli options
#  - select domain, ip, file
#  - includes errors
#  - etc

DOMAIN_FLAG = True
IP_FLAG = True
FILE_FLAG = True

URI = "http://[storage_ip]:[port]/samples/"
URI2 = "http://[storage_ip]:[port]/samples/"

RABBIT_USERNAME = 'guest'
RABBIT_PASSWORD = 'guest'
RABBIT_IP ='127.0.0.1'
RABBIT_PORT = 5672

credentials = pika.PlainCredentials(RABBIT_USERNAME, RABBIT_PASSWORD)
parameters = pika.ConnectionParameters(RABBIT_IP,
    5672,
    '/',
    credentials)
connection = pika.BlockingConnection(parameters)
channel = connection.channel()

channel.queue_declare(queue='totem_input',
    durable = True,
    exclusive = False,
    auto_delete = False,
    )

# TODO: add some more test cases. IPv6, domain with a subdomain, etc.
file_list = []
if IP_FLAG:
    file_list.append(("8.8.8.8", "ip"))
if DOMAIN_FLAG:
    file_list.append(("google.com", "domain"))
if FILE_FLAG:
    # We should change these to files that we can provide over github. 
    # Ma-shell has some binary files we could probably use.
    file_list.append(("0a4efbe854f1fa444303ca210842e779b55570216f62a4a406c89c564dabaf97", "file"))
    file_list.append(("2a43108a60fc5db94a001117093b6fff95697a8a5ab9a836b3c33c733c374c29", "file"))
    file_list.append(("04efbe854f1fa444303ca210842e779b55570216f62a4a406c89c564dabaf97", "file")) #fails
    file_list.append(("4b3287ef7e1b6add22621dcc97984e7aa12ca8ebb4625fc5eb36c9e707e4eb5f", "file"))
    file_list.append(("6fc9eb27fb82ead3a45a0fcee147eae01e12b9b36f587ac3e965d34b2ab59528", "file"))
    file_list.append(("7b365fe20f882ecce7096e56366df43c7e84f221902216c70b9ec3e2e68698e0", "file"))
    file_list.append(("973cfe3c16d97b37b044517531c43f87fe5e50dfb5726f59452040af1b724ee8", "file"))
    file_list.append(("c24bfff732cb7cf7d52c38dc92e3621aca781252157ab37093efba561579719d", "file"))
    file_list.append(("d2b41f09c3d48f808c5677331baa667ecf2f8b747517a993920c59dd14cbc695", "file"))
    file_list.append(("f8d3414d805dc45a2f50e7e933a450acbcb3416017358281a2a860d6cbea015c", "file"))
    file_list.append(("fa572834b690927a7a0eaffaac96f8de3aa1a5efd9dda585f0637d63db4b105f", "file"))

# TODO: once nofile is added to totem, we will need to fix the URI and adjust the download flag
for f in file_list:
    jsonmsg = {
        "primaryURI": URI + f[0],
        "secondaryURI": URI2 + f[0],
        "filename": f[0],
        "tags": ["test"],
        "attempts": 0,
        #"comment": "",
    }

    jsontask = {}
    if f[1] == "ip":
        jsontask["download"] = False
        jsontask["tasks"] = {
            "ASNMETA": []
        }
    elif f[1] == "domain":
        jsontask["download"] = False
        jsontask["tasks"] = {
            "DNSMETA": []
        }
    elif f[1] == "file":
        jsontask["download"] = True
        jsontask["tasks"] = {
            "GOGADGET": [],
            "OBJDUMP": [],
            "PEID": [],
            "PEINFO": [],
            "VIRUSTOTAL": [],
            "YARA": [],
        }

    jsonmsg.update(jsontask)
    msgBody = json.dumps(jsonmsg)
    channel.basic_publish(exchange='totem', routing_key='work.static.totem', body=msgBody)
    print(msgBody + "\n")

connection.close()
