# Traffic

**Quickstart**

````bash
git clone https://github.com/JackDanger/traffic.git
cd traffic
bin/traffic -f -c 5 fixtures/*.har
````

### HAR files and you: instant romance

Webkit can export network activity as a JSON expression of web requests
and responses called 'HAR' (HTTP Archive) files. These files capture all
of the exact requests your browser made and the responses they received
in an intuitive JSON format.

HAR files make it easy to replay the original request because they're
easily parsable by simple tools and contain (almost) all the information
you need to simulate real traffic.

### Love the PonyDebugger

PonyDebugger is a proxy (and iOS library) that captures network traffic
and shows it to you in the Webkit Web Inspector so you can have the same
powerful debugging tools as if you were making the requests in a
browser. This means that you can easily export an entire session of HTTP
requests from an iOS client.

Traffic is a tool for replaying HAR files to simulate load and to create
real-ish data. It executes the file as-is with a few possible
customizations:

#### GUIDs

You can specify GUIDs in the HAR that are realized at execution time. If
the string 'GUID1', 'GUID4', 'GUID78', etc. appear in your HAR file they will be
replaced by a session-consistent guid per-thread. So if you're, e.g.,
creating new objects and persisting them to the server this ensures
you're not running 10 threads all saving the same object repeatedly.

#### Concurrency

You can specify the level of concurrency under which to run your
archive(s). Any real site has multiple users operating multiple sessions
and your traffic simulation should reflect that.

#### Time-shifting

Expose bugs by playing the same HAR files faster or slower than they
were originally executed. Sometimes race conditions only appear when you
remove or greatly extend the time between two requests.

#### A Turing-complete config language

The reason HAR files aren't typically used is because there's no way to
connect important information from a response (say, a just-generated
session token) into the following request(s). Traffic provides a
regex-based system that lets you run arbitrary transformations of data
from any request or response to any subsequent one.

Pull requests welcome, forks celebrated.
