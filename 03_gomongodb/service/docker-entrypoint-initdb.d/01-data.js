db.orders.insertOne({
    _id: ObjectId('5f46f1c4c043dcee8f8e1062'),
    start: 1601571600000,
    film: {
        title: 'Неистовый',
        rating: 6.3,
        cashback: 0.15,
        genres: ['триллер'],
    },
    seats: [{row: 1, number: 3}, {row: 1, number: 4}],
    price: 200000,
    created: Date.now(),
});