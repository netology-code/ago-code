db.createUser({
    user: "app",
    pwd: "pass",
    roles: [
        {
            role: "readWrite",
            db: "db",
        }
    ]
});

db.createCollection('films');
