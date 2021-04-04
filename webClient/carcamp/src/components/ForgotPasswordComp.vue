<template>
<div class="forgotPassword">
    <h1>Forgot Password</h1>
    <div v-if="status=='username'">
      <p>Please enter your email address.</p>
      <form v-on:submit="doPasswordRest">
          <div class="form-group">
              <label for="exampleInputEmail1">Email address</label>
              <input type="email" class="form-control" id="exampleInputEmail1" aria-describedby="emailHelp" v-model="email">
              <small id="emailHelp" class="form-text text-muted">We'll never share your email with anyone else.</small>
          </div>
          <button type="submit" class="btn btn-primary">
              <div class="spinner-border" role="status" v-if="submitting">
                  <span class="sr-only">Loading...</span>
              </div>
              <span v-else>Submit</span> 
          </button>
      </form>
    </div>
    <div v-if="status=='reset'">
      <p>Please enter the code and the new password</p>
      <form v-on:submit="doConfirmPasswordReset">
          <div class="form-group">
              <label for="exampleInputEmail1">Code</label>
              <input type="text" class="form-control" id="exampleInputEmail1" aria-describedby="emailHelp" v-model="code">
              <small id="emailHelp" class="form-text text-muted">We'll never share your email with anyone else.</small>
          </div>
          <div class="form-group">
            <label for="exampleInputPassword1">New Password</label>
            <input type="password" class="form-control" id="exampleInputPassword1" v-model="newPassword">
        </div>
          <button type="submit" class="btn btn-primary">
              <div class="spinner-border" role="status" v-if="submitting">
                  <span class="sr-only">Loading...</span>
              </div>
              <span v-else>Submit</span> 
          </button>
      </form>
    </div>
</div>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import { mapActions } from 'vuex'

export default defineComponent({
  name: "ForgotPassword",
  props: {
    msg: String
  },
  data: function() {
    return {
      email: "",
      code: "",
      newPassword: "",
      submitting: false,
      status: "username"
    }
  },
  methods: {
    ...mapActions({
      passwordReset: "auth/passwordReset",
      confirmPasswordReset: "auth/confirmPasswordReset"
    }),
    doPasswordRest: async function(e) {
        e.preventDefault();
        this.submitting = true;
        await this.passwordReset({username: this.email});
        this.submitting = false;
        this.status = "reset";
    },
    doConfirmPasswordReset: async function(e) {
      e.preventDefault();
      this.submitting = true;
      await this.confirmPasswordReset({username: this.email, code: this.code, password: this.newPassword});
      this.submitting = false;
      this.$emit("did-reset-password");
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
