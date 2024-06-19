## url-shortener

Сократитель ссылок. 

### Микросовервисная архитектура
Отдельно выделен сервис бд для удобного масштабирование. Общаются через gRPC.
- [url-storage](https://github.com/quibex/url-storage)
