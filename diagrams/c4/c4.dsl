workspace "Name" "Description" {

    !identifiers hierarchical

    model {
        user = person "Пользователь"
        authUser = person "Авторизованный пользователь"
        admin = person "Администратор"

        app = softwareSystem "Airline tracker" {
            mobile = container "Мобильное приложение (Android)" {
                authController = component "Контроллер аутентификации"
                flightController = component "Контроллер полётов"

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

                flightRepository = component "Репозиторий полётов"{
                    tags "Database"
                }
                userRepository = component "Репозиторий пользователей"{
                    tags "Database"
                }
            }
        }

        user -> app.mobile "Просматривает полёты, используя"

        authUser -> app.mobile "Просматривает и отслеживает полёты, используя"

        admin -> app.mobile "Управляет полётами, используя"

        app.mobile -> app.backend "API"

        app.mobile.authController -> app.mobile.apiService "Использует"

        app.mobile.flightController -> app.mobile.flightRepository "Использует"

        app.mobile.apiService -> app.backend "API"

        app.mobile.flightRepository -> app.mobile.apiService "Использует"

        app.backend -> app.db "Хранит данные"
        app.backend -> app.cache "Хранит данные"

        app.mobile -> app.backend.adminController "Вызывает через HTTP"
        app.mobile -> app.backend.userController "Вызывает через HTTP"
        app.mobile -> app.backend.authController "Вызывает через HTTP"

        app.backend.adminController -> app.backend.authzService "Использует"
        app.backend.adminController -> app.backend.adminService "Использует"

        app.backend.userController -> app.backend.authzService "Использует"
        app.backend.userController -> app.backend.userService "Использует"

        app.backend.authController -> app.backend.authService "Использует"

        app.backend.authService -> app.backend.userRepository "Использует"

        app.backend.authzService -> app.backend.userRepository "Использует"

        app.backend.adminService -> app.backend.flightRepository "Использует"

        app.backend.userService -> app.backend.flightRepository "Использует"

        app.backend.flightRepository -> app.db "Читает/обновляет данные"
        app.backend.flightRepository -> app.cache "Читает/обновляет данные"

        app.backend.userRepository -> app.db "Читает/обновляет данные"
        app.backend.userRepository -> app.cache "Читает/обновляет данные"
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

        component app.mobile "Mobile-L3" {
            include *
            //include "->element.parent==app->"
            autolayout lr
        }

        component app.backend "Backend-L3" {
            include *
            //include "->element.parent==app->"
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
        }
    }
}