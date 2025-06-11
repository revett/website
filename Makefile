local:
	bunx @11ty/eleventy --input src --output dist --serve --port 5173 --watch --incremental

railway-build:
	bunx @11ty/eleventy --input src --output dist

railway-serve:
	bunx serve -l 8080 --no-port-switching --no-clipboard
