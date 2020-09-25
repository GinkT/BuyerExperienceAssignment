![alt text](https://i.imgur.com/D5h2K3Q.png "Архитектура сервиса")

НОРМАЛИЗОВАТЬ РАБОТУ МЕТОДА RUN, 
УДАЛЯТЬ ТОВАР ИЗ БД, В СЛУЧАЕ ОТСУТСТВИЯ, АДЕКВАТНУЮ СИСТЕМУ ЛОГОВ,
РАССМОТРЕТЬ МЬЮТЕКС ДЛЯ МАП, ФОРМУ ОТВЕТОВ - В JSON!
ДОБАВИТЬ ОПЦИИ КОНФИГУРАЦИИ,


# Пару слов о работе сервиса
Для ускорения доступа, сервис хранит данные о подписках/ценах товаров в оперативной памяти.       
При остановке/запуске сервиса происходит синхронизация с БД. Данные хранятся в двух структурах:

    var ProductSubs     map[ProductID][]string
    var ProductPrices   map[ProductID][]

_где ProductID - это пользовательский тип string, а []string в ProductSubs - email'ы пользователей_

В запущенном состоянии сервис раз в N секунд проходит по всей хэш таблице ProductSubs,  
применяет к её ключам(ID товаров) функцию GetCurrenctPrice и получает актуальную цену товара.   
В случае несовпадения с текущей ценой - обновляет цену и рассылает всем подписчикам товара письмо.
    

_Дата старта выполнения задания: 22.09.2020  
Дата окончания выполнения задания: 24.09.2020_