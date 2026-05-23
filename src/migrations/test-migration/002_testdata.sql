-- +goose Up

INSERT INTO departments (id, name, parent_id, created_at)
OVERRIDING SYSTEM VALUE VALUES
    -- depth 0: корни
    (1,  'Корпорация',      NULL, NOW() - INTERVAL '2 years'),
    (26, 'Холдинг',         NULL, NOW() - INTERVAL '2 years'),

    -- depth 1 под Корпорацией
    (2,  'Инженерия',       1,    NOW() - INTERVAL '2 years'),
    (3,  'Продукт',         1,    NOW() - INTERVAL '2 years'),
    (4,  'Операции',        1,    NOW() - INTERVAL '2 years'),

    -- depth 1 под Холдингом
    (27, 'Финансы',         26,   NOW() - INTERVAL '2 years'),
    (28, 'Маркетинг',       26,   NOW() - INTERVAL '2 years'),

    -- depth 2 под Инженерией
    (5,  'Бэкенд',          2,    NOW() - INTERVAL '20 months'),
    (6,  'Фронтенд',        2,    NOW() - INTERVAL '20 months'),
    (7,  'Инфраструктура',  2,    NOW() - INTERVAL '20 months'),
    (8,  'QA',              2,    NOW() - INTERVAL '18 months'),

    -- depth 2 под Продуктом
    (21, 'Аналитика',       3,    NOW() - INTERVAL '18 months'),
    (22, 'Дизайн',          3,    NOW() - INTERVAL '18 months'),

    -- depth 2 под Операциями
    (24, 'HR',              4,    NOW() - INTERVAL '18 months'),
    (25, 'Юридический',     4,    NOW() - INTERVAL '18 months'),

    -- depth 3 под Бэкендом
    (9,  'Go Team',         5,    NOW() - INTERVAL '16 months'),
    (10, 'Python Team',     5,    NOW() - INTERVAL '16 months'),
    (11, 'Java Team',       5,    NOW() - INTERVAL '16 months'),

    -- depth 3 под Фронтендом
    (12, 'React Team',      6,    NOW() - INTERVAL '15 months'),
    (13, 'Vue Team',        6,    NOW() - INTERVAL '15 months'),

    -- depth 3 под Инфраструктурой
    (14, 'DevOps',          7,    NOW() - INTERVAL '15 months'),
    (15, 'Security',        7,    NOW() - INTERVAL '15 months'),

    -- depth 3 под Дизайном
    (23, 'UX Research',     22,   NOW() - INTERVAL '12 months'),

    -- depth 4 под Go Team
    (16, 'Микросервисы',    9,    NOW() - INTERVAL '12 months'),
    (17, 'API Platform',    9,    NOW() - INTERVAL '12 months'),

    -- depth 4 под Python Team
    (18, 'ML Platform',     10,   NOW() - INTERVAL '10 months'),

    -- depth 4 под DevOps
    (19, 'Kubernetes',      14,   NOW() - INTERVAL '10 months'),

    -- depth 5 под Микросервисами — самый глубокий лист
    (20, 'Платёжный шлюз', 16,   NOW() - INTERVAL '6 months');

-- Сдвигаем sequence чтобы следующий автоинкремент не конфликтовал с явными id
ALTER TABLE departments ALTER COLUMN id RESTART WITH 100;

INSERT INTO employees (department_id, full_name, position, hired_at, created_at) VALUES
    -- Инженерия (2)
    (2,  'Алексей Иванов',     'Engineering Manager',     '2020-01-10', NOW() - INTERVAL '2 years'),
    (2,  'Марина Белова',      'Engineering Lead',        '2020-03-15', NOW() - INTERVAL '22 months'),

    -- Бэкенд (5)
    (5,  'Дмитрий Орлов',      'Tech Lead',               '2020-06-01', NOW() - INTERVAL '20 months'),
    (5,  'Елена Смирнова',     'Senior Developer',        '2021-01-20', NOW() - INTERVAL '18 months'),

    -- Go Team (9)
    (9,  'Сергей Попов',       'Senior Go Developer',     '2021-03-01', NOW() - INTERVAL '16 months'),
    (9,  'Анна Козлова',       'Go Developer',            '2022-05-10', NOW() - INTERVAL '14 months'),
    (9,  'Роман Фёдоров',      'Junior Go Developer',     NULL,         NOW() - INTERVAL '6 months'),

    -- Python Team (10)
    (10, 'Наталья Волкова',    'Python Lead',             '2021-07-15', NOW() - INTERVAL '15 months'),
    (10, 'Илья Морозов',       'Python Developer',        '2022-08-01', NOW() - INTERVAL '12 months'),

    -- Java Team (11)
    (11, 'Павел Лебедев',      'Java Developer',          '2022-02-14', NOW() - INTERVAL '13 months'),

    -- Фронтенд (6)
    (6,  'Ольга Новикова',     'Frontend Lead',           '2020-09-01', NOW() - INTERVAL '20 months'),

    -- React Team (12)
    (12, 'Кирилл Зайцев',      'React Developer',         '2022-04-01', NOW() - INTERVAL '14 months'),
    (12, 'Виктория Соколова',  'React Developer',         '2023-01-10', NOW() - INTERVAL '8 months'),

    -- Vue Team (13)
    (13, 'Максим Тихонов',     'Vue Developer',           NULL,         NOW() - INTERVAL '10 months'),

    -- DevOps (14)
    (14, 'Андрей Захаров',     'DevOps Engineer',         '2020-11-01', NOW() - INTERVAL '19 months'),
    (14, 'Светлана Громова',   'DevOps Engineer',         '2021-06-20', NOW() - INTERVAL '16 months'),

    -- Security (15)
    (15, 'Борис Лазарев',      'Security Engineer',       '2021-09-05', NOW() - INTERVAL '15 months'),

    -- Микросервисы (16)
    (16, 'Татьяна Егорова',    'Platform Engineer',       '2022-01-15', NOW() - INTERVAL '13 months'),
    (16, 'Николай Степанов',   'Backend Developer',       '2022-11-01', NOW() - INTERVAL '8 months'),

    -- ML Platform (18)
    (18, 'Юлия Кузнецова',     'ML Engineer',             '2022-03-01', NOW() - INTERVAL '12 months'),
    (18, 'Игорь Васильев',     'Data Scientist',          NULL,         NOW() - INTERVAL '5 months'),

    -- Платёжный шлюз (20) — depth 5
    (20, 'Людмила Орехова',    'Payment Systems Engineer','2023-02-01', NOW() - INTERVAL '7 months'),

    -- QA (8)
    (8,  'Владимир Соловьёв', 'QA Lead',                 '2021-04-01', NOW() - INTERVAL '17 months'),
    (8,  'Ирина Макарова',     'QA Engineer',             '2022-06-15', NOW() - INTERVAL '12 months'),

    -- Продукт (3)
    (3,  'Степан Григорьев',   'Product Director',        '2020-02-01', NOW() - INTERVAL '2 years'),

    -- Аналитика (21)
    (21, 'Валентина Михайлова','Business Analyst',        '2021-10-01', NOW() - INTERVAL '14 months'),

    -- UX Research (23)
    (23, 'Евгений Белоусов',   'UX Researcher',           NULL,         NOW() - INTERVAL '9 months'),

    -- HR (24)
    (24, 'Ксения Романова',    'HR Manager',              '2020-07-01', NOW() - INTERVAL '20 months'),

    -- Финансы (27) — дерево 2, цель для reassign
    (27, 'Антон Крылов',       'Finance Manager',         '2019-05-15', NOW() - INTERVAL '2 years'),
    (27, 'Марина Черных',      'Accountant',              '2020-08-20', NOW() - INTERVAL '19 months');

-- +goose Down
TRUNCATE TABLE employees, departments RESTART IDENTITY;
