(function() {
  const supportsCustomElements = 'customElements' in window;

  if (!supportsCustomElements) {
    document.body.innerHTML = "please upgrade your browser..."
  }


  class BaseElement extends HTMLElement {
    template = ``; // overwrite for each custom element

    constructor() {
      super();
      // console.log(this.tagName);
    }

    render(htmlstring) {
      if (!this.template) return this;
      let tpl = document.createElement('template');
      tpl.innerHTML = htmlstring || this.template;

      this.attachShadow({mode:'open'})
        .appendChild(tpl.content.cloneNode(true));
      return this;
    }

    fetch(url, args) {
      return window.fetch(url, args).then(r => r.json());
    }

    shadowQuery(qs) {
      return this.shadowRoot.querySelector(qs);
    }

    _ev_target(ev) {
      return ev.startsWith('G-') ? window : this;
    }

    emit(event, detail) {
      this._ev_target(event).dispatchEvent(new CustomEvent(event, { detail }));
      return this;
    }

    on(event, callback) {
      this._ev_target(event).addEventListener(event, callback);
      return this;
    }
  }

  window.customElements.define('dm-root', class extends BaseElement {
    template = `
      <style>
        :host {
          display: block;
          margin: 0 auto;
          width: 800px;
        }
      </style>
      <dm-header></dm-header>
      <dm-editor></dm-editor>
      <dm-body></dm-body>
      <dm-footer></dm-footer>`;

    connectedCallback() {
      this.render()
        .fetch('/api/index').then(d=> this.emit('G-fresh-index', d));
    }
  });

  window.customElements.define('dm-header', class extends BaseElement {
    template = `<style>
      :host h1 { margin-bottom:0; display: inline-block; }
      :host span { color #666; font-size:12px; }
      </style>
      <h1>Daymare</h1>
      <span>How Can a Daylight Know The Darkness Of Night</span>`;

    connectedCallback() {
      this.render()
    }
  });

  window.customElements.define('dm-footer', class extends BaseElement {
    template = `<style>
      :host p {
        --color: #ccc;
        text-align: right;
        color: var(--color);
        border-top: 1px solid var(--color);
      }
      :host img { width: 24px; vertical-align: bottom; }
      </style>
      <p>Power by
        <a href="http://github.com/yanyaoer/daymare" target="_blank">Project Daymare</a>
        <img src='favicon.svg' />
      </p>`

    connectedCallback() {
      this.render()
    }
  });

  window.customElements.define('dm-body', class extends BaseElement {
    connectedCallback() {
      this.render()
        .on('G-fresh-index', e => this.render_index(e.detail))
        .on('G-new-article', e => this.render_index(e.detail))
    }

    render_item(d) {
      const item = document.createElement('dm-article');
      const conv = new showdown.Converter({metadata: true});
      const html = conv.makeHtml(d.body)
      const meta = conv.getMetadata();
      item.innerHTML = `
        <h2 slot="title">
          ${meta.title || d.title}
          <dm-info>${meta.date || d.ctime}</dm-info>
        </h2>
        <div slot="body">${html}</div>`;
      return item;
    }

    render_index(data) {
      if (!data) return;
      data.forEach(d=> this.prepend(this.render_item(d)))
    }
  });

  window.customElements.define('dm-info', class extends BaseElement {
    template = `<style>
      :host { margin:0; color:#666; font-size: 11px; font-weight: normal; }
    </style>
    <slot></slot>`;

    connectedCallback() {
      this.render()
    }
  });

  window.customElements.define('dm-article', class extends BaseElement {
    // slotted element need this selector
    template = `<style>::slotted(h2) { margin: 1em 0 0; }</style>
    <div>
      <slot name="title"></slot>
      <slot name="body"></slot>
    </div>`;

    connectedCallback() {
      this.render()
    }
  });

  window.customElements.define('dm-editor', class extends BaseElement {
    template = `<style>
      :host textarea {
        display: block;
        width: 100%;
        height: 5em;
        background: #eee;
        border: 5px solid #ccc;
        border-sizing: border-box;
      }
    </style>
    <div>
      <form>
      <textarea name="body" placeholder="markdown support"></textarea>
      <div><button>submit</button></div>
      </form>
    </div>`;

    connectedCallback() {
      this.render().shadowQuery('form')
        .addEventListener('submit', ev => this.onSubmit(ev));
    }

    onSubmit(ev) {
      ev.preventDefault();
      const body = this.shadowQuery('textarea[name="body"]').value;
      this.fetch('/api/save', {
        method: 'POST',
        body: JSON.stringify({ body })
      }).then(d=> this.emit('G-new-article', [d]));
    }
  });

  document.body.appendChild(document.createElement('dm-root'));
})();
