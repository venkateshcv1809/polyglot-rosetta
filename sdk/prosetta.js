(function (global) {
  class ProsettaSDK {
    constructor() {
      this.initialized = false;
      this.version = "v1.0.0";
      this.baseUrl = "https://github.com/yourusername/polyglot-rosetta/releases/download";
    }

    async init(options = {}) {
      if (this.initialized) return;

      if (options.version) this.version = options.version;
      if (options.baseUrl) this.baseUrl = options.baseUrl;

      // 1. Inject wasm_exec.js if not already loaded
      if (typeof Go === "undefined") {
        await this._loadScript(`${this.baseUrl}/${this.version}/wasm_exec.js`);
      }

      // 2. Fetch and instantiate prosetta.wasm
      const wasmUrl = `${this.baseUrl}/${this.version}/prosetta.wasm`;
      const go = new Go();

      let inst;
      if ("instantiateStreaming" in WebAssembly) {
        const obj = await WebAssembly.instantiateStreaming(fetch(wasmUrl), go.importObject);
        inst = obj.instance;
      } else {
        const resp = await fetch(wasmUrl);
        const bytes = await resp.arrayBuffer();
        const obj = await WebAssembly.instantiate(bytes, go.importObject);
        inst = obj.instance;
      }

      // Run the Go WASM main routine in the background
      go.run(inst);
      this.initialized = true;
    }

    _loadScript(src) {
      return new Promise((resolve, reject) => {
        const script = document.createElement("script");
        script.src = src;
        script.onload = resolve;
        script.onerror = reject;
        document.head.appendChild(script);
      });
    }

    async run(concept, code) {
      if (!this.initialized) {
        throw new Error("Prosetta SDK not initialized. Call await Prosetta.init() first.");
      }
      // Delegates to exported Go WASM function exposed on window
      if (typeof global.wasmRunConcept !== "function") {
        throw new Error("WASM runtime methods not bound.");
      }
      return await global.wasmRunConcept(concept, code, false);
    }

    async submit(concept, code) {
      if (!this.initialized) {
        throw new Error("Prosetta SDK not initialized. Call await Prosetta.init() first.");
      }
      if (typeof global.wasmRunConcept !== "function") {
        throw new Error("WASM runtime methods not bound.");
      }
      return await global.wasmRunConcept(concept, code, true);
    }
  }

  global.Prosetta = new ProsettaSDK();
})(typeof window !== "undefined" ? window : globalThis);