-- ====================================================
-- 1. Вставка корневых подразделений (parent_id = NULL)
-- ====================================================
INSERT INTO department (name, parent_id) VALUES
    ('Корпоративный центр', NULL),
    ('Региональный филиал "Запад"', NULL),
    ('Региональный филиал "Восток"', NULL);

-- ====================================================
-- 2. Вставка дочерних подразделений (с привязкой к родителям через подзапросы)
-- ====================================================
-- Для корпоративного центра
INSERT INTO department (name, parent_id) VALUES
    ('Департамент разработки', (SELECT id FROM department WHERE name = 'Корпоративный центр' AND parent_id IS NULL)),
    ('Департамент маркетинга', (SELECT id FROM department WHERE name = 'Корпоративный центр' AND parent_id IS NULL)),
    ('Департамент продаж',    (SELECT id FROM department WHERE name = 'Корпоративный центр' AND parent_id IS NULL));

-- Для филиала "Запад"
INSERT INTO department (name, parent_id) VALUES
    ('Отдел ИТ',     (SELECT id FROM department WHERE name = 'Региональный филиал "Запад"' AND parent_id IS NULL)),
    ('Отдел кадров', (SELECT id FROM department WHERE name = 'Региональный филиал "Запад"' AND parent_id IS NULL));

-- Для филиала "Восток"
INSERT INTO department (name, parent_id) VALUES
    ('Отдел ИТ', (SELECT id FROM department WHERE name = 'Региональный филиал "Восток"' AND parent_id IS NULL)); -- имя "Отдел ИТ" разрешено, т.к. родитель другой

-- ====================================================
-- 3. Вставка подразделений второго уровня (внутри департаментов)
-- ====================================================
-- Внутри "Департамент разработки" (корпоративного центра)
INSERT INTO department (name, parent_id) VALUES
    ('Команда бэкенда',
        (SELECT id FROM department WHERE name = 'Департамент разработки'
         AND parent_id = (SELECT id FROM department WHERE name = 'Корпоративный центр' AND parent_id IS NULL))),
    ('Команда фронтенда',
        (SELECT id FROM department WHERE name = 'Департамент разработки'
         AND parent_id = (SELECT id FROM department WHERE name = 'Корпоративный центр' AND parent_id IS NULL))),
    ('Команда DevOps',
        (SELECT id FROM department WHERE name = 'Департамент разработки'
         AND parent_id = (SELECT id FROM department WHERE name = 'Корпоративный центр' AND parent_id IS NULL)));

-- Внутри "Департамент маркетинга"
INSERT INTO department (name, parent_id) VALUES
    ('Отдел контента',
        (SELECT id FROM department WHERE name = 'Департамент маркетинга'
         AND parent_id = (SELECT id FROM department WHERE name = 'Корпоративный центр' AND parent_id IS NULL))),
    ('Отдел рекламы',
        (SELECT id FROM department WHERE name = 'Департамент маркетинга'
         AND parent_id = (SELECT id FROM department WHERE name = 'Корпоративный центр' AND parent_id IS NULL)));

-- Внутри "Департамент продаж"
INSERT INTO department (name, parent_id) VALUES
    ('Отдел активных продаж',
        (SELECT id FROM department WHERE name = 'Департамент продаж'
         AND parent_id = (SELECT id FROM department WHERE name = 'Корпоративный центр' AND parent_id IS NULL))),
    ('Отдел работы с ключевыми клиентами',
        (SELECT id FROM department WHERE name = 'Департамент продаж'
         AND parent_id = (SELECT id FROM department WHERE name = 'Корпоративный центр' AND parent_id IS NULL)));

-- ====================================================
-- 4. Вставка сотрудников
-- ====================================================
-- Сотрудники корпоративного центра (разные департаменты)
INSERT INTO employee (department_id, full_name, position, hired_at) VALUES
    ((SELECT id FROM department WHERE name = 'Корпоративный центр' AND parent_id IS NULL), 'Иванов Иван Иванович', 'Генеральный директор', '2020-01-10'),

    ((SELECT id FROM department WHERE name = 'Департамент разработки' AND parent_id = (SELECT id FROM department WHERE name = 'Корпоративный центр' AND parent_id IS NULL)), 'Петров Пётр Петрович', 'Руководитель разработки', '2021-03-15'),
    ((SELECT id FROM department WHERE name = 'Команда бэкенда' AND parent_id IN (SELECT id FROM department WHERE name = 'Департамент разработки')), 'Сидоров Алексей Владимирович', 'Senior Backend Engineer', '2022-05-20'),
    ((SELECT id FROM department WHERE name = 'Команда бэкенда'), 'Кузнецова Мария Игоревна', 'Backend Developer', '2023-01-10'),
    ((SELECT id FROM department WHERE name = 'Команда фронтенда'), 'Николаев Дмитрий Сергеевич', 'Frontend Team Lead', '2021-11-01'),
    ((SELECT id FROM department WHERE name = 'Команда DevOps'), 'Васильев Олег Николаевич', 'DevOps инженер', '2022-09-12'),

    ((SELECT id FROM department WHERE name = 'Департамент маркетинга'), 'Михайлова Анна Андреевна', 'Директор по маркетингу', '2019-07-01'),
    ((SELECT id FROM department WHERE name = 'Отдел контента'), 'Егорова Екатерина Павловна', 'Контент-менеджер', '2021-02-18'),
    ((SELECT id FROM department WHERE name = 'Отдел рекламы'), 'Смирнов Артём Денисович', 'PPC специалист', '2022-11-05'),

    ((SELECT id FROM department WHERE name = 'Департамент продаж'), 'Тихонов Максим Викторович', 'Руководитель отдела продаж', '2018-04-22'),
    ((SELECT id FROM department WHERE name = 'Отдел активных продаж'), 'Козлова Елена Игоревна', 'Менеджер по продажам', '2020-10-14'),
    ((SELECT id FROM department WHERE name = 'Отдел работы с ключевыми клиентами'), 'Алексеев Владимир Сергеевич', 'Key Account Manager', '2019-12-03');

-- Сотрудники филиала "Запад"
INSERT INTO employee (department_id, full_name, position, hired_at) VALUES
    ((SELECT id FROM department WHERE name = 'Региональный филиал "Запад"' AND parent_id IS NULL), 'Новиков Павел Андреевич', 'Директор филиала', '2017-06-01'),
    ((SELECT id FROM department WHERE name = 'Отдел ИТ' AND parent_id = (SELECT id FROM department WHERE name = 'Региональный филиал "Запад"')), 'Морозов Дмитрий Алексеевич', 'Системный администратор', '2021-08-24'),
    ((SELECT id FROM department WHERE name = 'Отдел кадров' AND parent_id = (SELECT id FROM department WHERE name = 'Региональный филиал "Запад"')), 'Григорьева Светлана Петровна', 'HR-менеджер', '2022-02-11');

-- Сотрудники филиала "Восток"
INSERT INTO employee (department_id, full_name, position, hired_at) VALUES
    ((SELECT id FROM department WHERE name = 'Региональный филиал "Восток"' AND parent_id IS NULL), 'Соколов Илья Владимирович', 'Директор филиала', '2019-09-17'),
    ((SELECT id FROM department WHERE name = 'Отдел ИТ' AND parent_id = (SELECT id FROM department WHERE name = 'Региональный филиал "Восток"')), 'Белова Наталья Сергеевна', 'IT-специалист', '2022-04-09');

-- Несколько сотрудников с NULL hired_at (опционально)
INSERT INTO employee (department_id, full_name, position, hired_at) VALUES
    ((SELECT id FROM department WHERE name = 'Команда бэкенда'), 'Фёдоров Андрей Васильевич', 'Стажёр-разработчик', NULL),
    ((SELECT id FROM department WHERE name = 'Отдел активных продаж'), 'Павлова Ольга Алексеевна', 'Ассистент отдела продаж', NULL);

-- ====================================================
-- 5. Проверка (опционально) – вывести структуру подразделений
-- ====================================================
-- WITH RECURSIVE dept_tree AS (
--     SELECT id, name, parent_id, 0 AS depth FROM department WHERE parent_id IS NULL
--     UNION ALL
--     SELECT d.id, d.name, d.parent_id, dt.depth + 1
--     FROM department d
--     JOIN dept_tree dt ON d.parent_id = dt.id
-- )
-- SELECT * FROM dept_tree ORDER BY depth, id;
