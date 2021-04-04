<template>
<div class="spotDetail">

</div>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import { mapGetters, mapActions } from 'vuex'

export default defineComponent({
  name: "SpotDetail",
  props: {
    spotId: String
  },
  data: function() {
    return {
      email: "",
      password: "",
      submitting: false
    }
  },
  computed: {
    ...mapGetters({
      authenticatedUser: 'auth/authenticatedUser'
    }),
  },
  methods: {
    ...mapActions({
      signIn: "auth/signIn"
    }),
    doSignIn: async function(e) {
        e.preventDefault();
        this.submitting = true;
        await this.signIn({email: this.email, password: this.password});
        this.submitting = false;
        this.$emit("did-sign-in");
    }
  },
});
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
</style>
