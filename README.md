# networktest

A simple tool for network debugging.

Each running instance contains an HTTP client and server running on a given port. The
client will periodically connect to other running server instances and log whether the
attempt was successful and how long it took. The server simply echos whatever it
receives.