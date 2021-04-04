<template>
<div class="signup">
    <h1>Sign up</h1>
    <form v-on:submit="doSignUp" v-if="status == 'registration'">
        <div class="form-group">
            <label for="exampleInputEmail1">Email address</label>
            <input type="email" class="form-control" id="exampleInputEmail1" aria-describedby="emailHelp" v-model="email">
            <small id="emailHelp" class="form-text text-muted">We'll never share your email with anyone else.</small>
        </div>
        <div class="form-group">
            <label for="exampleInputPassword1">Password</label>
            <input type="password" class="form-control" id="exampleInputPassword1" v-model="password">
        </div>
        <button type="submit" class="btn btn-primary">
            <div class="spinner-border" role="status" v-if="submitting">
                <span class="sr-only">Loading...</span>
            </div>
            <span v-else>Submit</span> 
        </button>
    </form>
    <div v-else-if="status == 'confirmation'">
        <p>We sent you a confirmation code to your email, please confirm your email with the link we sent you.</p>
        <button type="button" class="btn btn-secondary mt-5" v-on:click="doConfirmResend">
            <div class="spinner-border" role="status" v-if="submitting">
                <span class="sr-only">Loading...</span>
            </div>
            <span v-else>Resend link</span> 
        </button>

    </div>
</div>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import { mapActions, mapGetters } from 'vuex'

export default defineComponent({
  name: "SignUpComp",
  props: {
    msg: String
  },
  data: function() {
    return {
      email: "",
      password: "",
      submitting: false,
      status: "registration"
    }
  },
  methods: {
    ...mapActions({
      signUp: "auth/signUp",
      confirmResend: "auth/confirmResend"
    }),
    doSignUp: async function(e) {
        e.preventDefault();
        this.submitting = true;
        await this.signUp({username: this.email, password: this.password});
        this.submitting = false;
        this.status = "confirmation";
    },
    doConfirmResend: async function() {
        this.submitting = true;
        await this.confirmResend({username: this.email});
        this.submitting = false;
    }
  },
  computed: {
    ...mapGetters({
      authenticationStatus: 'auth/authenticationStatus'
    }),
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
