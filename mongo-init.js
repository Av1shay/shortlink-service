db.getCollection("shortlinks").createIndex({
    key: "text"
},{
    unique: true
})

db.getCollection("counters").insert({
    _id: "shortlinkId",
    seq: 0
})