import { afterEach } from "vitest";
import { cleanup } from "@testing-library/react";

// Unmount React trees between tests.
afterEach(cleanup);

// jsdom does not implement matchMedia — provide a minimal stub so components
// that call useMediaQuery render without throwing.
if (!window.matchMedia) {
  window.matchMedia = (query: string) =>
    ({
      matches: false,
      media: query,
      onchange: null,
      addEventListener: () => {},
      removeEventListener: () => {},
      addListener: () => {},
      removeListener: () => {},
      dispatchEvent: () => false,
    }) as unknown as MediaQueryList;
}

// jsdom doesn't implement Element.scrollIntoView at all (unlike window.scrollTo,
// which it stubs as a harmless no-op) — components that call it on a ref (Timeline
// Week/Month's auto-scroll-to-today) throw without this.
if (!Element.prototype.scrollIntoView) {
  Element.prototype.scrollIntoView = function scrollIntoView() {};
}

// jsdom doesn't implement URL.createObjectURL/revokeObjectURL at all — Export's
// download-link flow (Blob -> object URL) throws without this.
if (!URL.createObjectURL) {
  URL.createObjectURL = () => "blob:mock-url";
}
if (!URL.revokeObjectURL) {
  URL.revokeObjectURL = () => {};
}

// jsdom's <dialog> lacks showModal/close — stub them for Dialog tests.
if (typeof HTMLDialogElement !== "undefined") {
  if (!HTMLDialogElement.prototype.showModal) {
    HTMLDialogElement.prototype.showModal = function showModal() {
      this.open = true;
    };
  }
  if (!HTMLDialogElement.prototype.close) {
    HTMLDialogElement.prototype.close = function close() {
      this.open = false;
      this.dispatchEvent(new Event("close"));
    };
  }
}
