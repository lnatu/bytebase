<template>
  <div class="mx-auto w-full max-w-sm">
    <div>
      <img class="h-12 w-auto" src="../../assets/logo.svg" alt="Bytebase" />
      <h2 class="mt-6 text-3xl leading-9 font-extrabold text-main">
        Forgot your password?
      </h2>
    </div>

    <div class="mt-8">
      <div class="mt-6">
        <label class="block text-base font-medium leading-5 text-control">
          Please contact your Bytebase Admin to reset your password.
        </label>
      </div>
    </div>

    <div class="mt-6 relative">
      <div class="absolute inset-0 flex items-center" aria-hidden="true">
        <div class="w-full border-t border-control-border"></div>
      </div>
      <div class="relative flex justify-center text-sm">
        <router-link to="/auth/signin" class="accent-link bg-white px-2">
          Return to Sign in
        </router-link>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
export default {
  name: "PasswordForgot",
  setup(props, ctx) {},
};
</script>
