# vacancies

***Парсинг вакансий по ключевому слову с данными авторизации***

Перед работой с приложением, необходимо посетить файл
**config/config.yml** и внести ключевое слово для поиска.

**docker build -t parser . && docker run --rm parser**     - запуск

НЕ СДЕЛАНО:
•	Поместить полученные данные в очередь RabbitMQ

•	Принять данные и записать в базу данных MSSQL.

•	Сделать парсинг страниц с вакансиями и страницы 
  непосредственно вакансии конкурентным

---------------------------------------------------------

Приложение 2 (Запрос данных и передача в очередь):

•	Необходимо создать приложение, которое будет получать название и ссылку со страницы вакансий по ключевому слову

Алгоритм работы:

•	Запросить данные авторизации у Приложения 1.

•	Проверить, что авторизация успешна и получить вакансии по ключевому слову из конфига (например “тест”)

•	Достать из вакансий название и ссылку

•	Поместить полученные данные в очередь RabbitMQ

•	Принять данные и записать в базу данных MSSQL.
