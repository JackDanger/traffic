# Traffic

**Quickstart**

````bash
git clone git@git.squareup.com:jackdanger/traffic.git
cd traffic
bundle
bin/traffic -f -c 5 archives/*.har
````

### HAR files and you: instant romance

Webkit can export network activity as a JSON expression of web requests
and responses called 'HAR' (HTTP Archive) files. These files are ripe
for replayability because they're easily parsable by simple tools and
contain (almost) all the information you need to reproduce the original
request.

### Love the PonyDebugger

PonyDebugger is a tool that proxies iOS network traffic through a proxy
server that exposes all the HTTP requests from the iOS devies in the
Webkit developer tools UI. This means that you can easily export an
entire session of HTTP requests from an iOS even if the requests
happened over SSL and you have no special permissions on the device itself.

Traffic is a script for replaying HAR files to simulate load and to
create real-ish data. It executes the file as-is with a few possible
customizations:

#### GUIDs

You can specify GUIDs in the HAR that are realized at execution time. If
the string 'GUID1', 'GUID4', etc. appear in your HAR file they will be
replaced by a session-consistent guid per-thread. So if you're, e.g.,
creating new objects and persisting them to the server this ensures
you're not running 10 threads all saving the same object repeatedly.

#### Concurrency

You can specify the level of concurrency under which to run your
archive(s). Any real site has multiple users operating multiple sessions
and your simulation testing should reflect that.

#### Time-shifting

Expose bugs by playing the same HAR files faster or slower than they
were originally executed. Sometimes race conditions only appear when you
remove or greatly extend the time between two requests.

#### A Turing-complete config language

The reason HAR files aren't typically used is because there's no way to
connect important information from a response (say, a just-generated
session token) into the following request(s). Traffic provides a Ruby
proc-based system that lets you run arbitrary transformations of data
from any request or response to any subsequent one.


Questions? Read the source. More questions? Ask jackdanger@
