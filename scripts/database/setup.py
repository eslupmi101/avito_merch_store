import os

from dotenv import load_dotenv
import psycopg2
import psycopg2.extras

load_dotenv()

DB_PARAMS = {
    "dbname": os.getenv("POSTGRES_DB"),
    "user": os.getenv("POSTGRES_USER"),
    "password": os.getenv("POSTGRES_PASSWORD"),
    "host": os.getenv("DB_HOST"),
    "port": os.getenv("DB_PORT"),
    "sslmode": os.getenv("DB_SSLMODE", "disable"),
}

TABLES = {
    "users": """
        CREATE TABLE IF NOT EXISTS users (
            id SERIAL PRIMARY KEY,
            username VARCHAR(100) NOT NULL UNIQUE,
            password TEXT NOT NULL,
            balance INT NOT NULL DEFAULT 0 CHECK (balance >= 0 AND balance <= 100000000)
        );
        CREATE INDEX IF NOT EXISTS idx_users_username ON users (username);
        CREATE INDEX IF NOT EXISTS idx_users_id ON users (id);
    """,
    "merch": """
        CREATE TABLE IF NOT EXISTS merch (
            id SERIAL PRIMARY KEY,
            name VARCHAR(200) NOT NULL UNIQUE,
            price INT NOT NULL CHECK (price > 0 AND price <= 100000000)
        );
        CREATE INDEX IF NOT EXISTS idx_merch_name ON merch (name);
    """,
    "transactions_table": """
        CREATE TABLE IF NOT EXISTS transactions (
            id SERIAL PRIMARY KEY,
            sender INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
            recipient INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
            amount INT NOT NULL CHECK (amount > 0 AND amount <= 100000000)
        );
        CREATE INDEX IF NOT EXISTS idx_transactions_sender ON transactions (sender);
        CREATE INDEX IF NOT EXISTS idx_transactions_recipient ON transactions (recipient);
    """,
    "merch_orders": """
        CREATE TABLE IF NOT EXISTS merch_orders (
            id SERIAL PRIMARY KEY,
            owner INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
            merch INT NOT NULL REFERENCES merch(id) ON DELETE CASCADE
        );
        CREATE INDEX IF NOT EXISTS idx_merch_orders_owner ON merch_orders (owner);
    """
}

MERCH_DATA = [
    ("t-shirt", 80),
    ("cup", 20),
    ("book", 50),
    ("pen", 10),
    ("powerbank", 200),
    ("hoody", 300),
    ("umbrella", 200),
    ("socks", 10),
    ("wallet", 50),
    ("pink-hoody", 500),
]


def connect_db(db_params):
    """Connects to the PostgreSQL database."""
    try:
        conn = psycopg2.connect(**db_params)
        conn.autocommit = True
        return conn
    except psycopg2.Error as e:
        print(f"Database connection error: {e}")
        exit(1)


def validate_tables(conn):
    """Checks if required tables exist in the database."""
    with conn.cursor(cursor_factory=psycopg2.extras.DictCursor) as cur:
        cur.execute(
            """
            SELECT table_name FROM information_schema.tables 
            WHERE table_schema='public';
            """
        )
        existing_tables = {row["table_name"] for row in cur.fetchall()}

        required_tables = set(TABLES.keys()) - {"transactions"}  # transactions is an ENUM, not a table

        missing_tables = required_tables - existing_tables

        if missing_tables:
            print(f"Missing tables: {', '.join(missing_tables)}")
            return False

        print("All required tables are present.")
        return True


def create_tables(conn):
    """Creates necessary tables and ENUM types."""
    with conn.cursor() as cur:
        for name, sql in TABLES.items():
            try:
                cur.execute(sql)
                print(f"Table {name} created (or already exists).")
            except psycopg2.Error as e:
                print(f"Error creating {name}: {e}")


def insert_merch_data(conn):
    """Inserts predefined merchandise data if not already present."""
    with conn.cursor() as cur:
        cur.execute("SELECT COUNT(*) FROM merch;")
        count = cur.fetchone()[0]
        if count == 0:
            cur.executemany(
                "INSERT INTO merch (name, price) VALUES (%s, %s) ON CONFLICT (name) DO NOTHING;",
                MERCH_DATA,
            )
            print("Merchandise data successfully inserted into `merch` table.")
        else:
            print("Merchandise data already exists, skipping insert.")


def setup_database(db_params):
    """
    Setup the database connection using the provided configuration parameters.
    
    Args:
        db_params (dict): A dictionary containing the database connection parameters.
            The dictionary should have the following keys:
            
            - 'dbname' (str): POSTGRES_DB.
            - 'user' (str): POSTGRES_USER.
            - 'password' (str): POSTGRES_PASSWORD.
            - 'host' (str): DB_HOST.
            - 'port' (str or int): DB_PORT.
            - 'sslmode' (str, optional): DB_SSLMODE.
    """
    conn = connect_db(db_params)

    print("Checking database structure...")
    if not validate_tables(conn):
        print("Database structure is incorrect. Trying to create missing tables...")
        create_tables(conn)

    print("Loading merchandise data...")
    insert_merch_data(conn)

    conn.close()
    print("Database setup completed successfully!")
    


def main():
    # setup by .env
    setup_database(DB_PARAMS)


if __name__ == "__main__":
    main()

