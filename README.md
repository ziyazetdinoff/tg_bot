# Обучающий телеграм чат бот

Бот написан на языке Golang и были применены следующая библиотека: GO Telegram Bot Api

## Функционал:
1) Напоминание об учёбе через определённый период (минута / час / день / неделя)
2) Скидывание ссылок на обучающие материалы из гугл таблицы

## Что нужно для запуска

Все необходимые для запуска переменные указаны в константах:
- `TELEGRAM_TOKEN` - токен телеграм бота, его можно получить с помощью телеграм бота @BotFather
- `GOOGLE_API_KEY` - апи токен от гугла, который нужен для взаимодействия с гугл таблицей. О том как его получить можно почитать здесь: https://developers.google.com/workspace/guides/create-credentials#api-key 
- `SPREADSHEET_ID` - id таблицы, он находится в ссылке на таблицу

![](screenshots/1.png)
![](screenshots/2.png)
![](screenshots/3.png)
![](screenshots/4.png)
