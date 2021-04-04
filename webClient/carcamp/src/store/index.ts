import { createStore } from 'vuex'


// Modules
import auth from "./modules/auth";

const debug = process.env.NODE_ENV !== "production";

const store = createStore({
    modules: {
        auth
    },
    strict: debug
  })

export default store;