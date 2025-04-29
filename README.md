# GoFlux - RSS/Atom Feed Aggregator

GoFlux is an RSS/Atom feed aggregator developed using the Go programming language. It allows users to follow their favorite content from various websites in a single platform. This project was created to improve and practice Go language skills.

## 🚀 Features

*   **User Management:** Create and manage users.
*   **Feed Management:** Add, list, and manage RSS/Atom feeds.
*   **Feed Following:** Allow users to follow feeds they are interested in.
*   **Content Aggregation:** Periodically checks feeds, fetches new posts, and saves them to the database.
*   **Database Integration:** Uses PostgreSQL with `sqlc` for database queries.
*   **Database Migrations:** Manages database schema using `goose`.
*   **Command Line Interface (CLI):** Provides basic functionalities through the command line.

## 🛠️ Technologies Used

*   **Language:** [Go](https://golang.org/)
*   **Database:** [PostgreSQL](https://www.postgresql.org/)
*   **DB Driver:** [github.com/lib/pq](https://github.com/lib/pq)
*   **SQL Compiler/Code Generator:** [sqlc](https://sqlc.dev/)
*   **DB Migration Tool:** [goose](https://github.com/pressly/goose)

## 🗄️ Database Schema

The project utilizes the following tables:

*   `users`: Stores user information.
*   `feeds`: Contains information about added RSS/Atom feeds.
*   `feed_follows`: Tracks which user follows which feed.
*   `posts`: Stores posts fetched from feeds.

Schema definitions are in `sql/schema/`, and queries are in `sql/queries/`.