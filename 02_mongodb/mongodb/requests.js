db.createCollection('orders');

// db.createCollection('orders', {
//     validator: {
//         $jsonSchema: {
//             bsonType: 'object',
//             required: ['film', 'start', 'duration', 'price', 'created'],
//             properties: {
//                 film: {
//                     bsonType: 'object',
//                     required: ['title'],
//                     properties: {
//                         title: {
//                             bsonType: 'string'
//                         }
//                     }
//                 },
//                 start: {
//                     bsonType: 'date'
//                 }
//             }
//         }
//     }
// });

db.orders.insertOne({
    start: 1601571600000,
    film: {
        title: 'Неистовый',
        rating: 6.3,
        cashback: 0.15,
        genres: ['триллер'],
    },
    seats: [{row: 1, number: 3}, {row: 1, number: 4}],
    price: 200000,
    created: new Date(),
});

db.orders.find();

db.orders.find({price: 200000});

db.orders.find({price: {$gt: 100000}});

db.orders.find({'film.rating': {$gt: 5.0}});

db.orders.find({film: {title: 'Неистовый'}});

db.orders.find({'film.genres': {$in: ['триллер', 'боевик']}});


db.orders.find({}, {film: {genres: 0}});

db.orders.updateOne({}, {$set: {'film.cashback': 0.10}});

db.orders.updateOne({_id: ObjectId('5f322eee9eb9202499b59443')}, {
    $push: {seats: { $each: [{row: 8, number: 5}, {row: 8, number: 6}]}}
});


db.orders.updateOne({_id: ObjectId('5f322eee9eb9202499b59443')}, {
    $pull: {seats: {$or: [{ row: 8, number: 5}, {row: 8, number: 6}]}}
});

db.orders.find();

db.orders.deleteMany({});