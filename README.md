#GoGoMux - A Fash HTTP Request Router

GoGoMux is a high performance HTTP request router
(also called *multiplexer* or just *mux* for short) for [Go](http://golang.org/).

In contrast to the default mux of Go's net/http package, this router supports
variables in the routing pattern and matches against the request method.
It has near linear scaleing.

## Definition: Go-Go

According to my dictionary, "Aggressivly Dynamic".

## Why a replacment router

I have been using Gorilla Mux for some time.  I just can't live with the
slow routing performance.  I tried HttpRouter.  It claims to be the fastest
router available and it is wikid fast. However it lacks
support for a variety of routing patterns that I already have in 
production.   This router is as fast (in my use case faster) than
HttpRouter. 

I will be adding to this to make it a near
drop-in replacement for Gorilla Mux.

## Not ready yet.

This is the initial checkin and the router with benchmarks and 
compatibility tests is not ready yet.  In other words - at some point
you just have to  take the working code, check it in and start
validating things like git, website links etc.   

