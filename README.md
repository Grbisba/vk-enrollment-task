# vk-enrollment-task

# Описание сервиса

Данный сервис предоставляет возможность оформлять подписки. Есть возможность подписаться на события по ключу и опубликовать события по ключу для всех подписчиков.

# Содержание

- [Методы](#Методы)
- [Запуск сервера](#запуск-сервиса)
- [Ожидалось в решении](#ожидалось-в-решении)
- [Пакет «subpub»](#пакет-subpub)

## Методы

### gRPC Publish – Публикация сообщения.

Данный метод предоставляет возможность публиковать "data" в топик по его "key".

#### Формат запроса.
```protobuf
syntax="proto3";

message PublishRequest {
  string key = 1;
  string data = 2;
}
```

#### Формат ответа.

- Пустой protobuf в случае, если публикация завершилась успешно.
- InvalidArgument, в случае, если запрос равен nil.
- Internal, в случае, если произошла ошибка во время записи сообщения.

### gRPC Subscribe – Подписка на сервис.

Данный метод предоставляет возможность подписаться на топик с помощью "key". Если же топик со значением "key" не найден, то сервис создаст его и будет возвращать значение, как только "publisher" его опубликует в данный топик.

#### Формат запроса.
```protobuf
syntax="proto3";

message SubscribeRequest {
  string key = 1;
}
```

#### Формат ответа.

```protobuf
syntax="proto3";

message Event {
  string data = 1;
}
```

- Поток сообщений, если публикация завершилась успешно.
- InvalidArgument, в случае, если запрос равен nil.
- Internal, в случае, если произошла ошибка во время подписки.

## Запуск сервиса.

1) Склонируйте репозиторий.
```bash
git clone git@github.com:Grbisba/vk-enrollment-task.git
```
2) Создайте файл "config.json" в корневой директории проекта:
   
   Содержание файла "config.json":
```json
{
  "controller": {
    "host": "<укажите хост - string>",
    "port": "<укажите порт - int>",
    "timeout_seconds": "<укажите таймаут - int>"
  }
}

```
3) Запустите сборку и запуск контейнера.
```bash
make build
```

## Ожидалось в решении.

### Патерны, задействованные в разработке данного сервиса.
1) Graceful Shutdown – успешная остановка сервера, листенера и pubsub шины. 

   [Файл с реализацией](server/internal/server/controller/grpc/grpc.go)
2) Dependency Injection – реализован с помощью `go.uber.fx`.

   [Файл с реализацией](server/cmd/server/server.go)

### Логирование.
Логирование предоставляется с помощью создания кастомизированного zap.Logger.

[Файл с реализацией](server/pkg/logger/logger.go)

## Пакет «subpub»

Пакет «subpub» предоставляет возможность для работы с шиной событий, работающую по принципу pub-sub (Publisher-Subscriver).

Интерфейсы и функции для работы с "subpub" пакетом.

```go
package subpub

import (
	"context"
)

// MessageHandler is a callback function that processes messages delivered to subscribers.
type MessageHandler func(msg interface{})

type Subscription interface {
	// Unsubscribe will remove interest in the current subject subscription is for
	Unsubscribe()
}

type SubPub interface {
	// Subscribe creates an asynchronous queue subscriber on the given subject.
	Subscribe(subject string, cb MessageHandler) (Subscription, error)

	// Publish publishes the msg argument to the given subject.
	Publish(subject string, msg interface{}) error

	// Close will shut down a sub-pub system.
	// May be blocked by data delivery until the context is canceled.
	Close(ctx context.Context) error
}
```

### Функция «MessageHandler».

#### Функция «MessageHandler»

Функция типа MessageHandler, передаваемая в «Subscribe», для обработки сообщения «msg».

### Интерфейс Subscription.

#### Метод «Unsubscribe»

Метод «Unsubscribe» позволяет отписаться от события в шине. Удаляет событие конкретно для текущего подписчика, вычитывая все сообщения, отправленные в него.

### Интерфейс SubPub
#### Метод «Subscribe»

Метод «Subscribe» позволяет подписаться на событие в шине.

- «Subscribe» принимает два аргумента:
  - Строка «subject» – имя события, на которое необходимо подписаться.
  - Функция типа MessageHandler, передаваемая в «Subscribe», для обработки сообщения «msg».

- Возвращает:
  - Интерфейс "Subscription", который позволяет отменить подписку с помощью метода "Unsubscribe".
  - Ошибку "Error".

#### Метод «Publish»

Метод «Publish» позволяет отправить сообщение в шину событий.

- «Publish» принимает два аргумента:
   - Строка «subject» – имя события, в которое необходимо отправить сообщение.
   - Интерфейс «msg», являющийся сообщением, которое публикуют.

- Возвращает:
    - Ошибку "Error".

#### Метод «Close»

Метод «Close» позволяет закрыть шину событий.

- «Close» принимает один аргумент:
   - context.Context «ctx» – контекст, для отмены закрытия шины.

- Возвращает:
    - Ошибку "Error".