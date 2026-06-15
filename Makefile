.PHONY: build serve css thumbs clean

## build: compile CSS and render the site into public/
build: css
	hugo --minify

## css: compile Tailwind into static/style.css
css:
	tailwindcss -i assets/css/main.css -o static/style.css --minify

## serve: dev server on :1313 with live reload (CSS and Hugo both watch)
serve:
	tailwindcss -i assets/css/main.css -o static/style.css --watch & \
	hugo server ; \
	kill %1

## thumbs: regenerate the committed 320px post-cover thumbnails (macOS sips)
thumbs:
	for f in $$(grep -h '^cover:' content/posts/*.md | awk '{print $$2}'); do \
		sips -Z 320 "static$$f" --out "static/images/thumbs/$$(basename $$f)" >/dev/null; \
	done

## clean: remove build artefacts
clean:
	rm -rf public resources static/style.css
