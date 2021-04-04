<template>
  <div class="signin">
    <SignInComp v-if="state == 'signin'" v-on:did-sign-in="didSignIn" v-on:forgot-password="didForgetPassword" />
    <SignUpComp v-else-if="state == 'signup'" />
    <ForgotPassword v-else-if="state == 'forgot-password'" v-on:did-reset-password="didResetPassword" />

    <button type="button" class="btn btn-secondary mt-5" v-on:click="toggleSignup">
        <span v-if="state == 'signin' || state == 'forgot-password'">No Account? Sign Up</span>
        <span v-else-if="state == 'signup'">Have an Account? Sign In</span>
    </button>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import SignInComp from "@/components/SignInComp.vue"
import SignUpComp from "@/components/SignUpComp.vue"
import ForgotPassword from "@/components/ForgotPasswordComp.vue"

export default defineComponent({
  name: "SignIn",
  components: {
    SignInComp,
    SignUpComp,
    ForgotPassword
  },
  data: function() {
    return {
        state: "signin",
    }
  },
  methods: {
      didSignIn: function(){
          this.$router.push({name: "Home"})
      },
      didForgetPassword: function() {
          this.state = "forgot-password";
      },
      didResetPassword: function() {
        this.state = "signin";
      },
      toggleSignup: function() {
        if(this.state == "signin"){
          this.state = "signup";
        } else {
          this.state = "signin";
        }
      }
  }
});
</script>
