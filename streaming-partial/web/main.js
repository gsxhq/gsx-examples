// gsx dev panel: Cmd-D / Ctrl-D
import "virtual:gsx-devpanel";
import "./style.css";
import "./gsxui.css";
import "./gsxui/index.js";
import { setupCounter } from "./counter.js";

setupCounter(document.querySelector("#counter"));
