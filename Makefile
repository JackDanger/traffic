install_babel:
	npm install --save-dev babel-cli
	npm install --save-dev babel-preset-react

server/_site/app.js: server/_site/javascript.jsx
	node node_modules/babel-cli/bin/babel.js server/_site/javascript.jsx > server/_site/app.js

test:
	go run github.com/JackDanger/traffic/...

server/key.pem:
	./generate_ssl_cert.sh

server/cert.pem:
	./generate_ssl_cert.sh

server: install_babel server/_site/app.js server/key.pem
	go run main.go server
