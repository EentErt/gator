# Gator
## Requirements
Postgres and Go must be installed to run the program
## How to Install
### In a terminal, enter
`go install https://github.com/EentErt/gator`

### Create the config file
In your home directory, create a file called `.gatorconfig.json`
the config file should contain the following Json:
```
{
  "db_url":"postgres://postgres:postgres@localhost:5432/gator?sslmode=disable",
  "current_user_name":""
}
```
The current user name will be set by the program

## Running the program
Use `gator {command} {args}` to run the program.

Commands:
  register {name}: register a name to the database
  login {name}: login as a user
  users: get a list of users
  agg {time interval}: begin the aggregation loop. This will periodically fetch the feeds followed by the logged in user.
    For time interval use a simple string to represent the time: 15s (15 seconds), 5m (5 minutes), etc.
  addfeed {title} {url}: add a feed to the database
  feeds: get a list of the feeds in the table
  follow {feed title}: follow a feed
  following: get a list of followed feeds
  unfollow: unfollow a feed
  browse {number of posts}: look at a number of most recent posts. If no argument is given, show 2 posts.
  
