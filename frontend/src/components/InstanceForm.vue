<style scoped>
/*  Removed the ticker in the number field  */
input::-webkit-outer-spin-button,
input::-webkit-inner-spin-button {
  -webkit-appearance: none;
  margin: 0;
}

/* Firefox */
input[type="number"] {
  -moz-appearance: textfield;
}
</style>

<template>
  <form class="space-y-6 divide-y divide-block-border">
    <div class="space-y-6 divide-y divide-block-border px-1">
      <!-- Instance Name -->
      <div class="grid grid-cols-1 gap-y-6 gap-x-4 sm:grid-cols-3">
        <div class="sm:col-span-2 sm:col-start-1">
          <label for="environment" class="textlabel">
            Environment <span v-if="create" style="color: red">*</span>
          </label>
          <!-- Disallow changing environment after creation. This is to take the conservative approach to limit capability -->
          <EnvironmentSelect
            class="mt-1 w-full"
            id="environment"
            name="environment"
            :disabled="!create"
            :selectedId="
              create
                ? state.instance.environmentId
                : state.instance.environment.id
            "
            @select-environment-id="
              (environmentId) => {
                if (create) {
                  state.instance.environmentId = environmentId;
                } else {
                  updateInstance('environmentId', environmentId);
                }
              }
            "
          />
        </div>

        <div class="sm:col-span-2 sm:col-start-1">
          <label for="name" class="textlabel">
            Instance Name <span style="color: red">*</span>
          </label>
          <input
            required
            id="name"
            name="name"
            type="text"
            class="textfield mt-1 w-full"
            :disabled="!allowEdit"
            :value="state.instance.name"
            @input="updateInstance('name', $event.target.value)"
          />
        </div>

        <div class="sm:col-span-2 sm:col-start-1">
          <label for="host" class="textlabel block">
            Host or Socket <span style="color: red">*</span>
          </label>
          <input
            required
            type="text"
            id="host"
            name="host"
            placeholder="e.g. 127.0.0.1 | localhost | /tmp/mysql.sock"
            class="textfield mt-1 w-full"
            :disabled="!allowEdit"
            :value="state.instance.host"
            @input="updateInstance('host', $event.target.value)"
          />
        </div>

        <div class="sm:col-span-1">
          <label for="port" class="textlabel block"> Port </label>
          <input
            type="number"
            id="port"
            name="port"
            placeholder="3306"
            class="textfield mt-1 w-full"
            :disabled="!allowEdit"
            :value="state.instance.port"
            @input="updateInstance('port', $event.target.value)"
          />
        </div>

        <div class="sm:col-span-3 sm:col-start-1">
          <label for="externallink" class="textlabel inline-flex">
            <span class="">External Link</span>
            <button
              class="ml-1 btn-icon"
              :disabled="state.instance.externalLink?.trim().length == 0"
              @click.prevent="
                window.open(urlfy(state.instance.externalLink), '_blank')
              "
            >
              <svg
                class="w-4 h-4"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
                xmlns="http://www.w3.org/2000/svg"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"
                ></path>
              </svg>
            </button>
          </label>
          <input
            required
            id="externallink"
            name="externallink"
            type="text"
            class="textfield mt-1 w-full"
            :disabled="!allowEdit"
            :value="state.instance.externalLink"
            @input="updateInstance('externalLink', $event.target.value)"
          />
        </div>
      </div>
      <!-- Read/Write Connection Info -->
      <div class="pt-4">
        <div class="flex justify-between">
          <div>
            <h3 class="text-lg leading-6 font-medium text-gray-900">
              Connection info
            </h3>
            <p
              class="mt-1 text-sm text-gray-500"
              :class="create ? 'max-w-xl' : ''"
            >
              This is the connection used by Bytebase to perform DDL and DML
              operations. We recommend you to create a separate user
              'bytebase@%' to operate. Below is an example to create such an
              user and grant it with the needed privileges.
            </p>
            <div class="mt-2 text-sm text-main">
              <div class="flex flex-row">
                <span
                  class="
                    flex-1
                    min-w-0
                    w-full
                    inline-flex
                    items-center
                    px-3
                    py-2
                    border border-r border-control-border
                    bg-gray-50
                    sm:text-sm
                    whitespace-pre
                  "
                >
                  {{ GRANT_STATEMENT }}
                </span>
                <button
                  tabindex="-1"
                  class="
                    -ml-px
                    px-2
                    py-2
                    border border-gray-300
                    text-sm
                    font-medium
                    text-control-light
                    disabled:text-gray-300
                    bg-gray-50
                    hover:bg-gray-100
                    disabled:bg-gray-50
                    focus:ring-control focus:outline-none
                    focus-visible:ring-2
                    focus:ring-offset-1
                    disabled:cursor-not-allowed
                  "
                  @click.prevent="copyGrantStatement"
                >
                  <svg
                    class="w-6 h-6"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                    xmlns="http://www.w3.org/2000/svg"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"
                    ></path>
                  </svg>
                </button>
              </div>
            </div>
          </div>
        </div>
        <div class="pt-4 grid grid-cols-1 gap-y-4 gap-x-4 sm:grid-cols-3">
          <div class="sm:col-span-1 sm:col-start-1">
            <label for="username" class="textlabel block"> Username </label>
            <!-- For mysql, username can be empty indicating anonymous user. 
            But it's a very bad practice to use anonymous user for admin operation,
            thus we make it REQUIRED here. -->
            <input
              id="username"
              name="username"
              type="text"
              class="textfield mt-1 w-full"
              :disabled="!allowEdit"
              :value="state.instance.username"
              @input="state.instance.username = $event.target.value"
            />
          </div>

          <div class="sm:col-span-1 sm:col-start-1">
            <label for="password" class="textlabel block"> Password </label>
            <input
              id="password"
              name="password"
              type="text"
              class="textfield mt-1 w-full"
              autocomplete="off"
              placeholder="sensitive - write only"
              :disabled="!allowEdit"
              :value="create ? state.instance.password : state.updatedPassword"
              @input="
                create
                  ? (state.instance.password = $event.target.value)
                  : (state.updatedPassword = $event.target.value)
              "
            />
          </div>
        </div>
        <div v-if="showTestConnection" class="pt-8 space-y-2">
          <div class="flex flex-row space-x-2">
            <button
              @click.prevent="testConnection"
              type="button"
              class="btn-normal whitespace-nowrap items-center"
            >
              Test Connection
            </button>
          </div>
        </div>
      </div>
    </div>
    <!-- Action Button Group -->
    <div class="pt-4">
      <!-- Create button group -->
      <div v-if="create" class="flex justify-end">
        <button
          type="button"
          class="btn-normal py-2 px-4"
          @click.prevent="cancel"
        >
          Cancel
        </button>
        <button
          type="button"
          class="btn-primary ml-3 inline-flex justify-center py-2 px-4"
          :disabled="!allowCreate"
          @click.prevent="tryCreate"
        >
          Create
        </button>
      </div>
      <!-- Update button group -->
      <div v-else class="flex justify-end">
        <button
          v-if="allowEdit"
          type="button"
          class="btn-normal ml-3 inline-flex justify-center py-2 px-4"
          :disabled="!valueChanged"
          @click.prevent="doUpdate"
        >
          Update
        </button>
      </div>
    </div>
  </form>
  <BBAlert
    v-if="state.showCreateInstanceWarningModal"
    :style="'WARN'"
    :okText="'Ignore and create'"
    :title="'Connection info seems to be incorrect'"
    :description="state.createInstanceWarning"
    @ok="
      () => {
        state.showCreateInstanceWarningModal = false;
        doCreate();
      }
    "
    @cancel="state.showCreateInstanceWarningModal = false"
  >
  </BBAlert>
</template>

<script lang="ts">
import { computed, reactive, PropType, ComputedRef } from "vue";
import { useStore } from "vuex";
import { useRouter } from "vue-router";
import cloneDeep from "lodash-es/cloneDeep";
import isEqual from "lodash-es/isEqual";
import { toClipboard } from "@soerenmartius/vue3-clipboard";
import EnvironmentSelect from "../components/EnvironmentSelect.vue";
import { instanceSlug, isDBAOrOwner } from "../utils";
import {
  Instance,
  InstanceCreate,
  UNKNOWN_ID,
  Principal,
  InstancePatch,
  ConnectionInfo,
  SqlResultSet,
} from "../types";
import isEmpty from "lodash-es/isEmpty";

const GRANT_STATEMENT =
  "CREATE USER bytebase@'%' IDENTIFIED BY '{{YOUR_SECRET_PWD}}';\n\nGRANT ALTER, ALTER ROUTINE, CREATE, CREATE ROUTINE, CREATE USER, \nCREATE VIEW, DELETE, DROP, EXECUTE, INDEX, PROCESS, REFERENCES, \nSELECT, SHOW DATABASES, SHOW VIEW, TRIGGER, UPDATE, USAGE \nON *.* to bytebase@'%';";

interface LocalState {
  originalInstance?: Instance;
  instance: Instance | InstanceCreate;
  // Only used in non-create case.
  updatedPassword: string;
  showCreateInstanceWarningModal: boolean;
  createInstanceWarning: string;
}

export default {
  name: "DataSourceCreateForm",
  emits: ["dismiss"],
  props: {
    create: {
      default: "false",
      type: Boolean,
    },
    instance: {
      // Can be false when create is true
      required: false,
      type: Object as PropType<Instance>,
    },
  },
  components: { EnvironmentSelect },
  setup(props, { emit }) {
    const store = useStore();
    const router = useRouter();

    const currentUser: ComputedRef<Principal> = computed(() =>
      store.getters["auth/currentUser"]()
    );

    const state = reactive<LocalState>({
      originalInstance: props.instance,
      // Make hard copy since we are going to make equal comparsion to determine the update button enable state.
      instance: props.instance
        ? cloneDeep(props.instance)
        : {
            environmentId: UNKNOWN_ID,
            name: "New Instance",
            engine: "MYSQL",
            host: "127.0.0.1",
            username: "",
          },
      updatedPassword: "",
      showCreateInstanceWarningModal: false,
      createInstanceWarning: "",
    });

    const allowCreate = computed(() => {
      return state.instance.name && state.instance.host;
    });

    const allowEdit = computed(() => {
      return (
        props.create ||
        ((state.instance as Instance).rowStatus == "NORMAL" &&
          isDBAOrOwner(currentUser.value.role))
      );
    });

    const showTestConnection = computed(() => {
      return (
        props.create ||
        ((state.instance as Instance).rowStatus == "NORMAL" &&
          isDBAOrOwner(currentUser.value.role))
      );
    });

    const valueChanged = computed(() => {
      return (
        !isEqual(state.instance, state.originalInstance) ||
        !isEmpty(state.updatedPassword)
      );
    });

    const updateInstance = (field: string, value: string) => {
      (state.instance as any)[field] = value;
    };

    const cancel = () => {
      emit("dismiss");
    };

    const tryCreate = () => {
      const connectionInfo: ConnectionInfo = {
        dbType: "MYSQL",
        username: state.instance.username,
        password: state.instance.password,
        host: state.instance.host,
        port: state.instance.port,
      };
      store
        .dispatch("sql/ping", connectionInfo)
        .then((resultSet: SqlResultSet) => {
          if (isEmpty(resultSet.error)) {
            doCreate();
          } else {
            state.createInstanceWarning = `Bytebase is unable to connect the instance. We recommend you to review the connection info again. But it's OK to ignore this warning for now. You can still fix the connection info from the instance detail page after creation.\n\nError detail: ${resultSet.error}`;
            state.showCreateInstanceWarningModal = true;
          }
        });
    };

    // We will also create the database * denoting all databases
    // and its RW data source. The username, password is actually
    // stored in that data source object instead of in the instance self.
    // Conceptually, data source is the proper place to store connnection info (thinking of DSN)
    const doCreate = () => {
      store
        .dispatch("instance/createInstance", state.instance)
        .then((createdInstance) => {
          emit("dismiss");

          router.push(`/instance/${instanceSlug(createdInstance)}`);

          store.dispatch("notification/pushNotification", {
            module: "bytebase",
            style: "SUCCESS",
            title: `Successfully created instance '${createdInstance.name}'.`,
          });

          // After creating the instance, we will check if migration schema exists on the instance.
          setTimeout(() => {}, 1000);
        });
    };

    const doUpdate = () => {
      const patchedInstance: InstancePatch = {};
      if (state.instance.name != state.originalInstance!.name) {
        patchedInstance.name = state.instance.name;
      }
      if (state.instance.externalLink != state.originalInstance!.externalLink) {
        patchedInstance.externalLink = state.instance.externalLink;
      }
      if (state.instance.host != state.originalInstance!.host) {
        patchedInstance.host = state.instance.host;
      }
      if (state.instance.port != state.originalInstance!.port) {
        patchedInstance.port = state.instance.port;
      }
      if (state.instance.username != state.originalInstance!.username) {
        patchedInstance.username = state.instance.username;
      }
      if (!isEmpty(state.updatedPassword)) {
        patchedInstance.password = state.updatedPassword;
      }

      store
        .dispatch("instance/patchInstance", {
          instanceId: (state.instance as Instance).id,
          instancePatch: patchedInstance,
        })
        .then((instance) => {
          state.originalInstance = instance;
          // Make hard copy since we are going to make equal comparsion to determine the update button enable state.
          state.instance = cloneDeep(state.originalInstance!);
          state.updatedPassword = "";

          store.dispatch("notification/pushNotification", {
            module: "bytebase",
            style: "SUCCESS",
            title: `Successfully updated instance '${instance.name}'.`,
          });
        });
    };

    const copyGrantStatement = () => {
      toClipboard(GRANT_STATEMENT).then(() => {
        store.dispatch("notification/pushNotification", {
          module: "bytebase",
          style: "INFO",
          title: `CREATE USER and GRANT statements copied to clipboard. Paste to your mysql client to apply.`,
        });
      });
    };

    const testConnection = () => {
      const connectionInfo: ConnectionInfo = {
        dbType: "MYSQL",
        username: state.instance.username,
        password: props.create
          ? state.instance.password
          : state.updatedPassword,
        host: state.instance.host,
        port: state.instance.port,
        instanceId: props.create ? undefined : (state.instance as Instance).id,
      };
      store
        .dispatch("sql/ping", connectionInfo)
        .then((resultSet: SqlResultSet) => {
          if (isEmpty(resultSet.error)) {
            store.dispatch("notification/pushNotification", {
              module: "bytebase",
              style: "SUCCESS",
              title: `Successfully connected instance.`,
            });
          } else {
            store.dispatch("notification/pushNotification", {
              module: "bytebase",
              style: "CRITICAL",
              title: `Failed to connect instance.`,
              description: resultSet.error,
              // Manual hide, because user may need time to inspect the error
              manualHide: true,
            });
          }
        });
    };

    return {
      GRANT_STATEMENT,
      state,
      allowCreate,
      allowEdit,
      showTestConnection,
      valueChanged,
      updateInstance,
      cancel,
      tryCreate,
      doCreate,
      doUpdate,
      copyGrantStatement,
      testConnection,
    };
  },
};
</script>
