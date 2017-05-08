install_babel:
	npm install --save-dev babel-cli
	npm install --save-dev babel-preset-react

compile_jsx:
	node node_modules/babel-cli/bin/babel.js server/_site/javascript.jsx > server/_site/app.js

test:
	go run github.com/JackDanger/traffic/...

server: install_babel compile_jsx
	go run main.go server
