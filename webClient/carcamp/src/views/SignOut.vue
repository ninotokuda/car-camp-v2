<template>
  <div class="signout">
    <h1>Sign out</h1>
    <button type="button" class="btn btn-primary" v-on:click="doSignOut">
        <div class="spinner-border" role="status" v-if="submitting">
            <span class="sr-only">Loading...</span>
        </div>
        <span v-else>Signout</span> 
    </button>
    
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import { mapActions } from 'vuex'

export default defineComponent({
  name: "SignOut",
  data: function() {
    return {
      submitting: false
    }
  },
  methods: {
    ...mapActions({
      signOut: "auth/signOut"
    }),
    doSignOut: async function(e) {
        e.preventDefault();
        this.submitting = true;
        await this.signOut();
        this.submitting = false;
        this.$router.push({name: "Home"});
    }
  },
});
</script>
