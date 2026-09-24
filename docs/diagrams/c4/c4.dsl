workspace "Name" "Description" {

    !identifiers hierarchical

    model {
        user = person "Пользователь"
        authUser = person "Авторизованный пользователь"
        admin = person "Администратор"
        mailServer = softwareSystem "Почтовый сервер"

        app = softwareSystem "Airline tracker" {
            web = container "Сайт" {
                authView = component "Экран аутентификации"
                flightView = component "Экран с полётами"

                authViewModel = component "Контроллер аутентификации"
                flightViewModel = component "Контроллер полётов"

                authModel = component "Сервис аутентификации"
                flightModel = component "Сервис полётов"

                apiService = component "Сервис API"

                flightRepository = component "Репозиторий полётов"{
                    tags "Database"
                }
            }
            db = container "БД" {
                tags "Database"
            }
            cache = container "Кэш" {
                tags "Database"
            }
            backend = container "Бэкенд" {
                adminController = component "Контроллер администратора"
                userController = component "Контроллер пользователя"
                authController = component "Контроллер аутентификации"

                authService = component "Сервис аутентификации"
                authzService = component "Сервис авторизации"
                adminService = component "Сервис администратора"
                userService = component "Сервис пользователя"
                notificationService = component "Сервис уведомлений"

                flightRepository = component "Репозиторий полётов"{
                    tags "Database"
                }
                userRepository = component "Репозиторий пользователей"{
                    tags "Database"
                }
            }
            notifier = container "Сервис отправки уведомлений" {
                eventReceiver = component "Сервис уведомлений"
                eventHandlers = component "Обработчик событий "
                notificationRepository = component "Репозиторий уведомлений"
                notificationSender = component "Планировщик отправки"
                emailService = component "Сервис email-уведомлений"
                smtpClient = component "SMTP-клиент"
            }
            kafka = container "Kafka" {
                tags "Kafka"
            }
        }

        user -> app.web "Просматривает полёты, используя"

        authUser -> app.web "Просматривает и отслеживает полёты, используя"

        admin -> app.web "Управляет полётами, используя"

        app.web -> app.backend "API"

        app.web.authView -> app.web.authViewModel "Вызывает"
        app.web.flightView -> app.web.flightViewModel "Вызывает"

        app.web.authViewModel -> app.web.authModel "Вызывает"
        app.web.flightViewModel -> app.web.flightModel "Вызывает"

        app.web.authModel -> app.web.apiService "Использует"

        app.web.flightModel -> app.web.flightRepository "Использует"

        app.web.apiService -> app.backend "API"

        app.web.flightRepository -> app.web.apiService "Использует"

        app.backend -> app.db "Хранит данные"
        app.backend -> app.cache "Хранит данные"
        app.backend -> app.kafka "Отправляет сообщения"

        app.web -> app.backend.adminController "Вызывает через HTTP"
        app.web -> app.backend.userController "Вызывает через HTTP"
        app.web -> app.backend.authController "Вызывает через HTTP"

        app.backend.adminController -> app.backend.authzService "Использует"
        app.backend.adminController -> app.backend.adminService "Использует"

        app.backend.userController -> app.backend.authzService "Использует"
        app.backend.userController -> app.backend.userService "Использует"
        app.backend.userController -> app.backend.notificationService "Использует"

        app.backend.authController -> app.backend.authService "Использует"
        app.backend.authService -> app.backend.userRepository "Использует"

        app.backend.authzService -> app.backend.userRepository "Использует"

        app.backend.adminService -> app.backend.flightRepository "Использует"

        app.backend.userService -> app.backend.flightRepository "Использует"

        app.backend.notificationService -> app.backend.userRepository "Использует"
        app.backend.notificationService -> app.backend.flightRepository "Использует"
        app.backend.notificationService -> app.kafka "Отправляет данные"

        app.backend.flightRepository -> app.db "Читает/обновляет данные"
        app.backend.flightRepository -> app.cache "Читает/обновляет данные"

        app.backend.userRepository -> app.db "Читает/обновляет данные"
        app.backend.userRepository -> app.cache "Читает/обновляет данные"

        app.notifier -> app.kafka "Получает данные"
        app.notifier.eventReceiver -> app.kafka "Получает события" "Kafka"
        app.notifier.eventReceiver -> app.notifier.eventHandlers "Передаёт событие"

        app.notifier.eventHandlers -> app.notifier.notificationRepository "Сохраняет сформированные уведомления"

        app.notifier.notificationSender -> app.notifier.notificationRepository "Получает неотправленные уведомления и изменяет их статус"
        app.notifier.notificationSender -> app.notifier.emailService "Передаёт уведомление для отправки"

        app.notifier.emailService -> app.notifier.smtpClient "Передаёт сформированное письмо"
        app.notifier.smtpClient -> mailServer "Отправляет письмо" "SMTP"

        app.notifier.notificationRepository -> app.db "Читает и записывает уведомления" "SQL"
        mailServer -> authUser "Доставляет уведомление по электронной почте
    }

    views {
        systemContext app "L1" {
            include *
            autolayout lr
        }

        container app "L2" {
            include *
            autolayout lr
        }

        component app.web "Web-L3" {
            include *
            //include "->element.parent==app->"
            autolayout lr
        }

        component app.backend "Backend-L3" {
            include *
            //include "->element.parent==app->"
            autolayout lr
        }

        component app.notifier "Notifier-L3" {
            include *
            include app.kafka
            include app.db
            include mailServer
            include authUser
            autolayout lr
        }

        styles {
            element "Element" {
                color #9a28f8
                stroke #9a28f8
                strokeWidth 7
                shape roundedbox
            }
            element "Person" {
                shape person
            }
            element "Database" {
                shape cylinder
            }
            element "Boundary" {
                strokeWidth 5
            }
            relationship "Relationship" {
                thickness 4
            }
            element "Kafka" {
                shape pipe
            }
        }
    }
}
