import htmx from "htmx.org";
import Alpine from "alpinejs";

// Add Alpine and htmx instance to window object.
window.htmx = htmx;
window.Alpine = Alpine;

// Start Alpine.
Alpine.start();
