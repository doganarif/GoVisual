import { h, render } from "preact";
import { App } from "./App";

// Mount the app
const root = document.getElementById("app");
if (root) {
  render(<App />, root);
} else {
  console.error("Could not find app root element");
}
