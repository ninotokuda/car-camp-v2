import Auth from "@aws-amplify/auth";
import findIndex from 'lodash/findIndex';

// initial state
const state = {
  user: null,
  userSession: null,
  isAuthenticated: false,
  authenticationStatus: null,
  authenticationError: null
};

const getters = {
  authenticatedUser: state => state.user,
  isAuthenticated: state => state.isAuthenticated,
  authenticationStatus: state => {
    return state.authenticationStatus
      ? state.authenticationStatus
      : { variant: "secondary" };
  },
  hasAuthenticationStatus: state => {
    return !!state.authenticationStatus;
  },
  authenticatedUserIsAdmin: state => {
    try {
			const groups = state.userSession.idToken.payload["cognito:groups"];
			return findIndex(groups, (e) => e == "Admin") != -1;
    } catch {
      return false
    }
  },
	authenticatedUserIdToken: state => {
		try {
			return state.userSession.idToken.jwtToken;
		} catch {
			return null
		}
	},
	authenticatedUserId: state => {
		try {
			return state.userSession.idToken.payload["cognito:username"];
		} catch {
			return null
		}
  },
  authenticationError: state => state.authenticationError
};

const mutations = {
  setAuthenticationError(state, err) {
    state.authenticationError = err;
    state.authenticationStatus = {
      state: "failed",
      message: err.message,
      variant: "danger"
    };
  },
  clearAuthenticationStatus: state => {
    state.authenticationStatus = null;
    state.authenticationError = null;
  },
  setUserAuthenticated(state, userSession) {
    state.userSession = userSession;
    state.isAuthenticated = true;
  },
  clearAuthentication(state) {
    state.user = null;
    state.userId = null;
    state.userSession = null;
    state.isAuthenticated = false;
  }
};

const actions = {
  clearAuthenticationStatus: context => {
    context.commit("clearAuthenticationStatus", null);
  },
  signIn: async (context, params) => {
    context.commit("auth/clearAuthenticationStatus", null, { root: true });
    try {
      const user = await Auth.signIn(params.email, params.password);
      context.commit("setUserAuthenticated", user.signInUserSession);
    } catch (err) {
      context.commit("auth/setAuthenticationError", err, { root: true });
    }
  },
  signOut: async context => {
    try {
      await Auth.signOut();
    } catch (err) {
      context.commit("auth/setAuthenticationError", err, { root: true });
    }
    context.commit("auth/clearAuthentication", null, { root: true });
  },
  signUp: async (context, params) => {
    context.commit("auth/clearAuthenticationStatus", null, { root: true });
    try {
      await Auth.signUp(params);
      context.commit("auth/clearAuthentication", null, { root: true });
    } catch (err) {
      context.commit("auth/setAuthenticationError", err, { root: true });
    }
  },
  confirmSignUp: async (context, params) => {
    context.commit("auth/clearAuthenticationStatus", null, { root: true });
    try {
      await Auth.confirmSignUp(params.username, params.code);
    } catch (err) {
      context.commit("auth/setAuthenticationError", err, { root: true });
    }
  },
  confirmResend: async (context, params) => {
    context.commit("auth/clearAuthenticationStatus", null, { root: true });
    try {
      await Auth.resendSignUp(params.username);
    } catch (err) {
      context.commit("auth/setAuthenticationError", err, { root: true });
    }
  },
  passwordReset: async (context, params) => {
    context.commit("auth/clearAuthenticationStatus", null, { root: true });
    try {
      await Auth.forgotPassword(params.username);
    } catch (err) {
      context.commit("auth/setAuthenticationError", err, { root: true });
    }
  },
  confirmPasswordReset: async (context, params) => {
    context.commit("auth/clearAuthenticationStatus", null, { root: true });
    try {
      await Auth.forgotPasswordSubmit(
        params.username,
        params.code,
        params.password
      );
    } catch (err) {
      context.commit("auth/setAuthenticationError", err, { root: true });
    }
  },
  fetchCurrentSession: async context => {
    context.commit("auth/clearAuthenticationStatus", null, { root: true });
    try {
      const userSession = await Auth.currentSession();
      context.commit("setUserAuthenticated", userSession);
    } catch (err) {
      context.commit("auth/setAuthenticationError", err, { root: true });
    }
  },
};

export default {
  namespaced: true,
  state,
  getters,
  actions,
  mutations
};