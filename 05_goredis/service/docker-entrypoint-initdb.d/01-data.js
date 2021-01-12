db.films.insertMany([
    {
        _id: ObjectId('5f46f1c4c043dcee8f8e1061'),
        title: 'Вратарь Галактики',
        rating: 4.7,
        cashback: 0.15,
        genres: ['детский', 'семейный', 'приключения', 'фантастика'],
        start: Date.now(),
    },
    {
        _id: ObjectId('5f46f1c4c043dcee8f8e1062'),
        title: 'Довод',
        rating: 8.0,
        cashback: 0.15,
        genres: ['триллер', 'драма', 'боевик'],
        start: Date.now(),
    },
    {
        _id: ObjectId('5f46f1c4c043dcee8f8e1063'),
        title: 'Новые мутанты',
        rating: 5.7,
        cashback: 0.15,
        genres: ['фантастика', 'экшен'],
        start: Date.now(),
    }
]);
