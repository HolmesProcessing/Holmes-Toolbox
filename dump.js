//mongo crits --quiet dump.js > out.txt

cursor = db.sample.find(
{
    $or: [
// Apply your filters
//        {'source.name': 'ii'},
        {'source.name': 'maltrieve'},
//        {'source.name': 'novetta'},
//        {'source.name': 'I20'},
//        {'source.name': 'yaraexchange'},
//        {'source.name': 'virusshare'}
    ]
},
    {
        _id: 1,
    }
);

while(cursor.hasNext()) {
    printjsononeline(cursor.next());
}
