<template>
<div class="edit">
    <h1>Edit Spot</h1>
    <form v-on:submit="doSignIn">
        <div class="form-group">
            <label for="exampleInputEmail1">Email address</label>
            <input type="email" class="form-control" id="exampleInputEmail1" aria-describedby="emailHelp" v-model="email">
            <small id="emailHelp" class="form-text text-muted">We'll never share your email with anyone else.</small>
        </div>
        <div class="form-group">
            <label for="exampleInputPassword1">Password</label>
            <input type="password" class="form-control" id="exampleInputPassword1" v-model="password">
            <button type="button" class="btn btn-link" v-on:click="$emit('forgot-password')">Forgot your Password?</button>
        </div>
        <button type="submit" class="btn btn-primary">
            <div class="spinner-border" role="status" v-if="submitting">
                <span class="sr-only">Loading...</span>
            </div>
            <span v-else>Submit</span> 
        </button>
    </form>
</div>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import { mapGetters, mapActions } from 'vuex'

export default defineComponent({
  name: "EditSpot",
  props: {
    msg: String
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
h3 {
  margin: 40px 0 0;
}
ul {
  list-style-type: none;
  padding: 0;
}
li {
  display: inline-block;
  margin: 0 10px;
}
a {
  color: #42b983;
}
</style>
