import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import test from "node:test";
import vm from "node:vm";

const config = readFileSync(new URL("../../frontend/src/js/config.js", import.meta.url), "utf8");
const frontend = readFileSync(new URL("../../frontend/src/index.html", import.meta.url), "utf8");
const documentation = readFileSync(new URL("../../backend/api/cmd/docs/index.html", import.meta.url), "utf8");

for (const prefix of ["", "/mirror", "/reserved"]) {
  test(`frontend URLs under ${prefix || "/"}`, () => {
    const origin = "http://localhost:8080";
    const documentURL = new URL(`${origin}${prefix}/`);
    for (const asset of ["styles.css", "app.js"]) {
      const match = frontend.match(new RegExp(`(?:href|src)="([^\"]*${asset.replace(".", "\\.")})"`));
      assert.ok(match, `missing ${asset}`);
      assert.equal(new URL(match[1], documentURL).pathname, `${prefix}/${asset}`);
    }

    // Evaluate the browser module using its deployed URL, without an HTTP server.
    const moduleURL = `${origin}${prefix}/js/config.js`;
    const source = config.replace(/^export /gm, "")
      .replaceAll("import.meta.url", JSON.stringify(moduleURL));
    const apiBase = vm.runInNewContext(`${source}\nAPI_BASE`, { URL });
    assert.equal(apiBase, `${prefix}/api/v1`);
  });
}

for (const prefix of ["", "/mirror"]) {
  for (const suffix of ["", "index.html"]) {
    test(`Swagger UI URLs under ${prefix || "/"}, ${suffix || "directory"}`, () => {
      const location = new URL(`http://127.0.0.1:8080${prefix}/api/v1/docs/${suffix}`);
      const window = { location };
      let options;
      const SwaggerUIBundle = (value) => {
        options = value;
        return {};
      };
      SwaggerUIBundle.presets = { apis: {} };
      const script = documentation.match(/<script>\s*([\s\S]*?)<\/script>/)[1];
      vm.runInNewContext(script, { URL, window, SwaggerUIBundle });
      window.onload();

      assert.equal(options.url, `${location.origin}${prefix}/api/v1/openapi.yml`);
      for (const asset of ["swagger-ui.css", "swagger-ui-bundle.js"]) {
        const match = documentation.match(new RegExp(`(?:href|src)="([^\"]*${asset.replaceAll(".", "\\.")})"`));
        assert.equal(new URL(match[1], location).pathname, `${prefix}/api/v1/docs/${asset}`);
      }

      const request = { url: "http://localhost:8080/api/v1/flights?limit=5&cursor=abc" };
      options.requestInterceptor(request);
      assert.equal(request.url, `${location.origin}${prefix}/api/v1/flights?limit=5&cursor=abc`);

      const alreadyPrefixed = { url: `${location.origin}/mirror/api/v1/flights` };
      options.requestInterceptor(alreadyPrefixed);
      assert.equal(alreadyPrefixed.url, `${location.origin}/mirror/api/v1/flights`);
    });
  }
}
