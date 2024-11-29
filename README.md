# gator

A blog feed aggregator command line tool.

## Usage

The intended usage for this program (as of this version) is to run `gator agg
and let the aggregator run in the background.  In a new terminal, login and browse
the articles available in the followed RSS feeds.

## Installation

To install the program, there are a few prerequisites.

- Postgres
- Go
- Goose

When Postgres is installed and you have a connection string, you need to do two things:

- Create a config file
  Create a file called `.gatorconfig.json` in your home directory, with the format

```json

{
  "db_url": "postgres://user:password@server:port/gator?sslmode=disable"
}

```

  This will tell `gator` how to connect to your database.

- Run the goose migrations
  Navigate to the `sql/schema` directory and run the command
`goose postgres <connection string> up`
This will create the proper tables in the database for the tool to run.

Then it should be ready to go!

## Commands

Current commands include:

- Login
  - Usage: `gator login <username>`
  Logs in a registered user.  This changes the feeds displayed when browsing articles.
- Register
  - Usage: `gator register <username>`
  Registers a new user and makes them available for login.
- Users
  - Usage: `gator users`
  Lists all registered users.
- Aggregate
  - Usage: `gator agg`
  Fetches and stores article data from RSS feeds.
- Add Feed
  - Usage: `gator addfeed <feed name> <url>`
  Adds a feed to the aggregator.  This also marks the user as following the feed
  that they have added.
- Follow
  - Usage: `gator follow <feed url>`
  If a feed has already been added to the database, this command will allow
  the logged-in user to follow that feed.
- Following
  - usage: `gator following`
  Lists the feeds the current user is following.
- Unfollow
  - usage: `gator unfollow <feed url>`
  Unfollows the feed; the posts will stop appearing for that user.
- Browse
  - usage: `gator browse <# articles (optional)>`
  Lists a selection of articles from the feeds the user is following.
  By default, two articles are displayed, but more can be shown with the argument.
