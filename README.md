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

# How GoGoMux came to be

Like most things, nesesity is the mother of invention.   My problem
was performance.   On a reasonably large set of routes I was seeing
between 10 and 15 pages served per second.   This truly horrified
me.   The serve was in Go and I was expecting sto see serveral
hunder per second.   Some testing quicly lead to the conclusion
that over half of the time was being used up by Gorilla Mux and its
processing.

I have been tuning and improving the performance of software for a
long time (I started programming on paper punch cards).  My first
rule of improving performace is you alwasy start with the biggest
time using item.  In this case it was figuring out what was going
on with Gorilla Mux and why it was taking so much itme.   A search
in the code lead me to the fact that it apears to use a linear
search algorythm with a set of regular expressions.  Since I had
over 100 routes and many of them used regular expressions this was
an important finding.   My first move was to take the route that
was most common and move it to the top of the search.  This one
change iproved my results from the 10 to 15 range to 80 to 100 per
second.   This change also gave me the breathing room and time to
look into the problem.

A quick search on the web revealed that HttpRouter was a lot faster.
Many thousands of tiems faster than Gorilla Mux.    Now was the
time for a quick code switch.   It used a signinficanly different
interface than Gorilla.  This added many hundereds of lines of code.
That took a few hours to get changed.  And... And...

HttpRouter only allows for unique routes.  My client code (some of
witch I did not write and did not want to change) uses some non-unique
routes.  My frustration level was going back up!  I went and looked
at the benchmarks that HttpRouter had in its documentation.  They
clearly implied that HttpRouter covered all routes for sites like
GitHub.   Then I doug into the code and the author had commented
out about 10% of the GitHub routes so that the reamining 90% where
unique.

I could now stick with HttpRoute and make major changes to my server
to add in a post-route disambiguation phase - or - I could dig into
the problem of implementing a fast router that was as close to
Gorilla Mux as possible.

I chose to implement a new router.   At the time I did not have a
clear understading of what the routing problem really involved.  It
seems so simple.  You just take a URL and map that string to a
function call.  I implemented that.  It was fast.  Not as fast as
HttpRoute but a lot faster than Gorilla Mux.   Then I really doug
into the test cases and discoverd that my first router would not
cut it.  It was not even close.  So a massive amount of chagnes and
a new version.  Then version 3, 4, 5.  Version 5 worked.   The code
afer 5 rounds of slash and burn without a clear understanding of
the problem was so ugly it was guranteed to cause nightmairs.  But
it workd and it was as fast as HttpRouter.

The current version is my 6th attempt at getting this all to work.
It is reasonably close to what Gorilla Mux implements.  There are
still some missing features that I am working on.  Without calls
to middleware it is close to as fast as HttpRouter.    Testing the
original Gorilla Mux code and my latest version of GoGoMux on an
set of 5000 routes from my production server indicates that GoGoMux
is around 120,000 times as fast as Gorilla Mux.   I will pull
together a set of benchmarks that clearly shows this and also modify
the very comprehensive HttpRouter benchmarks to include GoGoMux.

GoGoMux still needs some work.  I have started to use it in my
production server.  In a couple of weeks, with a buch more tests
and some clean up it should be ready for prime time.

