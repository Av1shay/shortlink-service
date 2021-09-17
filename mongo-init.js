db.getCollection("shortlinks").createIndex({
    key: "text"
},{
    unique: true,
    name: "shortlink_key"
})