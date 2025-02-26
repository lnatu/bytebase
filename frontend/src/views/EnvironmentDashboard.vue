<template>
  <div>
    <BBTab
      :tabItemList="tabItemList"
      :selectedIndex="state.selectedIndex"
      :reorderModel="state.reorder ? 'ALWAYS' : 'NEVER'"
      @reorder-index="reorderEnvironment"
      @select-index="selectEnvironment"
    >
      <BBTabPanel
        v-for="(item, index) in environmentList"
        :key="item.id"
        :active="index == state.selectedIndex"
      >
        <div v-if="state.reorder" class="flex justify-center pt-5">
          <button
            type="button"
            class="btn-normal py-2 px-4"
            @click.prevent="discardReorder"
          >
            Cancel
          </button>
          <button
            type="submit"
            class="btn-primary ml-3 inline-flex justify-center py-2 px-4"
            :disabled="!orderChanged"
            @click.prevent="doReorder"
          >
            Apply Change
          </button>
        </div>
        <EnvironmentDetail
          v-else
          :environmentSlug="environmentSlug(item)"
          @archive="doArchive"
        />
      </BBTabPanel>
    </BBTab>
  </div>
  <BBModal
    v-if="state.showCreateModal"
    :title="'Create Environment'"
    @close="state.showCreateModal = false"
  >
    <EnvironmentForm
      :create="true"
      :environment="DEFAULT_NEW_ENVIRONMENT"
      @create="doCreate"
      @cancel="state.showCreateModal = false"
    />
  </BBModal>

  <BBAlert
    v-if="state.showGuide"
    :style="'INFO'"
    :okText="'Do not show again'"
    :cancelText="'Dismiss'"
    :title="'How to setup \'Environment\' ?'"
    :description="'Each environment maps to one of your testing, staging, prod environment respectively.\n\nEnvironment is a global setting, one Bytebase deployment only contains a single set of environments.\n\nDatabase instances are created under a particular environment.'"
    @ok="
      () => {
        doDismissGuide();
      }
    "
    @cancel="state.showGuide = false"
  >
  </BBAlert>
</template>

<script lang="ts">
import {
  onMounted,
  onUnmounted,
  computed,
  reactive,
  watch,
  ComputedRef,
} from "vue";
import { useStore } from "vuex";
import { useRouter } from "vue-router";
import { array_swap } from "../utils";
import EnvironmentDetail from "../views/EnvironmentDetail.vue";
import EnvironmentForm from "../components/EnvironmentForm.vue";
import { Environment, EnvironmentCreate, Principal } from "../types";
import { BBTabItem } from "../bbkit/types";

const DEFAULT_NEW_ENVIRONMENT: EnvironmentCreate = {
  name: "New Env",
  approvalPolicy: "MANUAL_APPROVAL_ALWAYS",
};

interface LocalState {
  reorderedEnvironmentList: Environment[];
  selectedIndex: number;
  showCreateModal: boolean;
  reorder: boolean;
  showGuide: boolean;
}

export default {
  name: "EnvironmentDashboard",
  components: {
    EnvironmentDetail,
    EnvironmentForm,
  },
  props: {},
  setup(props, ctx) {
    const store = useStore();
    const router = useRouter();

    const currentUser: ComputedRef<Principal> = computed(() =>
      store.getters["auth/currentUser"]()
    );

    const state = reactive<LocalState>({
      reorderedEnvironmentList: [],
      selectedIndex: -1,
      showCreateModal: false,
      reorder: false,
      showGuide: false,
    });

    const selectEnvironmentOnHash = () => {
      if (environmentList.value.length > 0) {
        if (router.currentRoute.value.hash) {
          for (let i = 0; i < environmentList.value.length; i++) {
            if (
              environmentList.value[i].id ==
              router.currentRoute.value.hash.slice(1)
            ) {
              selectEnvironment(i);
              break;
            }
          }
        } else {
          selectEnvironment(0);
        }
      }
    };

    onMounted(() => {
      store.dispatch("command/registerCommand", {
        id: "bb.environment.create",
        registerId: "environment.dashboard",
        run: () => {
          createEnvironment();
        },
      });
      store.dispatch("command/registerCommand", {
        id: "bb.environment.reorder",
        registerId: "environment.dashboard",
        run: () => {
          startReorder();
        },
      });

      selectEnvironmentOnHash();

      if (!store.getters["uistate/introStateByKey"]("guide.environment")) {
        setTimeout(() => {
          state.showGuide = true;
          store.dispatch("uistate/saveIntroStateByKey", {
            key: "environment.visit",
            newState: true,
          });
        }, 1000);
      }
    });

    onUnmounted(() => {
      store.dispatch("command/unregisterCommand", {
        id: "bb.environment.create",
        registerId: "environment.dashboard",
      });
      store.dispatch("command/unregisterCommand", {
        id: "bb.environment.reorder",
        registerId: "environment.dashboard",
      });
    });

    watch(
      () => router.currentRoute.value.hash,
      () => {
        if (router.currentRoute.value.name == "workspace.environment") {
          selectEnvironmentOnHash();
        }
      }
    );

    const environmentList = computed(() => {
      return store.getters["environment/environmentList"]();
    });

    const tabItemList = computed((): BBTabItem[] => {
      if (environmentList) {
        const list = state.reorder
          ? state.reorderedEnvironmentList
          : environmentList.value;
        return list.map((item: Environment, index: number) => {
          return {
            title: (index + 1).toString() + ". " + item.name,
            id: item.id,
          };
        });
      }
      return [];
    });

    const createEnvironment = () => {
      stopReorder();
      state.showCreateModal = true;
    };

    const doCreate = (newEnvironment: EnvironmentCreate) => {
      store
        .dispatch("environment/createEnvironment", newEnvironment)
        .then(() => {
          state.showCreateModal = false;
          selectEnvironment(environmentList.value.length - 1);
        });
    };

    const doDismissGuide = () => {
      store.dispatch("uistate/saveIntroStateByKey", {
        key: "guide.environment",
        newState: true,
      });
      state.showGuide = false;
    };

    const startReorder = () => {
      state.reorderedEnvironmentList = [...environmentList.value];
      state.reorder = true;
    };

    const stopReorder = () => {
      state.reorder = false;
      state.reorderedEnvironmentList = [];
    };

    const reorderEnvironment = (sourceIndex: number, targetIndex: number) => {
      array_swap(state.reorderedEnvironmentList, sourceIndex, targetIndex);
      selectEnvironment(targetIndex);
    };

    const orderChanged = computed(() => {
      for (let i = 0; i < state.reorderedEnvironmentList.length; i++) {
        if (
          state.reorderedEnvironmentList[i].id != environmentList.value[i].id
        ) {
          return true;
        }
      }
      return false;
    });

    const discardReorder = () => {
      stopReorder();
    };

    const doReorder = () => {
      store
        .dispatch(
          "environment/reorderEnvironmentList",
          state.reorderedEnvironmentList
        )
        .then(() => {
          stopReorder();
        });
    };

    const doArchive = (environment: Environment) => {
      if (environmentList.value.length > 0) {
        selectEnvironment(0);
      }
    };

    const selectEnvironment = (index: number) => {
      state.selectedIndex = index;
      router.replace({
        name: "workspace.environment",
        hash: "#" + environmentList.value[index].id,
      });
    };

    const tabClass = computed(() => "w-1/" + environmentList.value.length);

    return {
      DEFAULT_NEW_ENVIRONMENT,
      state,
      environmentList,
      tabItemList,
      createEnvironment,
      doCreate,
      doArchive,
      doDismissGuide,
      reorderEnvironment,
      orderChanged,
      discardReorder,
      doReorder,
      selectEnvironment,
      tabClass,
    };
  },
};
</script>
