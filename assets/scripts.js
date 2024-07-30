import htmx from "htmx.org";
import Alpine from "alpinejs";
import _hyperscript from "hyperscript.org";

// Add Alpine and htmx instance to window object.
window.htmx = htmx;
window.Alpine = Alpine;

// Start Alpine and Hyperscript
Alpine.start();
_hyperscript.browserInit();
