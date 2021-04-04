import { createApp } from "vue";
import App from "./App.vue";
import router from "./router";
import store from "./store";
import Auth from "@aws-amplify/auth";
import config from "@/aws-exports";
import "mapbox-gl/dist/mapbox-gl.css";
import 'bootstrap/dist/css/bootstrap.css'


Auth.configure(config);
createApp(App)
  .use(router)
  .use(store)
  .mount("#app");
