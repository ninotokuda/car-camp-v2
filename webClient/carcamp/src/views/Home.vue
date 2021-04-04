<template>
  <div class="home">
    <img alt="Vue logo" src="../assets/logo.png" />
    <router-link to="SignOut" v-if="isAuthenticated">Logout</router-link>
    <router-link to="SignIn" v-else>Login</router-link>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import { mapGetters, mapActions } from 'vuex'

export default defineComponent({
  name: "Home",
  data: function() {
    return {}
  },
  computed: {
    ...mapGetters({
      isAuthenticated: 'auth/isAuthenticated'
    }),
  },
  methods: {
    ...mapActions({
      fetchCurrentSession: 'auth/fetchCurrentSession'
    })
  },
  created: async function() {
     console.log("created");
     await this.fetchCurrentSession();
     console.log("created2", this.isAuthenticated);
  }
});
</script>
