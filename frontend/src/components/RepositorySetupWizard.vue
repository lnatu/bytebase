<template>
  <div>
    <div class="textinfolabel">
      If you encounter errors during the process, please refer to our<a
        href="https://docs.bytebase.com/use-bytebase/vcs-integration/link-repository?ref=console"
        target="_blank"
        class="normal-link"
      >
        detailed guide</a
      >.
    </div>
    <BBStepTab
      class="pt-4"
      :stepItemList="stepList"
      :allowNext="allowNext"
      @try-change-step="tryChangeStep"
      @try-finish="tryFinishSetup"
      @cancel="cancel"
    >
      <template v-slot:0="{ next }">
        <RepositoryVCSProviderPanel :config="state.config" @next="next()" />
      </template>
      <template v-slot:1="{ next }">
        <RepositorySelectionPanel :config="state.config" @next="next()" />
      </template>
      <template v-slot:2>
        <RepositoryConfigPanel :config="state.config" />
      </template>
    </BBStepTab>
  </div>
</template>

<script lang="ts">
import { reactive } from "@vue/reactivity";
import { computed, PropType } from "@vue/runtime-core";
import { useStore } from "vuex";
import { useRouter } from "vue-router";
import isEmpty from "lodash-es/isEmpty";
import { v4 as uuidv4 } from "uuid";
import { BBStepTabItem } from "../bbkit/types";
import RepositoryVCSProviderPanel from "./RepositoryVCSProviderPanel.vue";
import RepositorySelectionPanel from "./RepositorySelectionPanel.vue";
import RepositoryConfigPanel from "./RepositoryConfigPanel.vue";
import {
  Project,
  ProjectRepositoryConfig,
  RepositoryCreate,
  unknown,
  VCS,
} from "../types";
import { projectSlug, randomString } from "../utils";

const CHOOSE_PROVIDER_STEP = 0;
const CHOOSE_REPOSITORY_STEP = 1;
const CONFIGURE_DEPLOY_STEP = 2;

interface LocalState {
  config: ProjectRepositoryConfig;
  currentStep: number;
}

const stepList: BBStepTabItem[] = [
  { title: "Choose Git provider", hideNext: true },
  { title: "Select repository", hideNext: true },
  { title: "Configure deploy" },
];

export default {
  name: "RepositorySetupWizard",
  emits: ["cancel", "finish"],
  components: {
    RepositoryVCSProviderPanel,
    RepositorySelectionPanel,
    RepositoryConfigPanel,
  },
  props: {
    // If false, then we intend to change the existing linked repository intead of just linking a new repository.
    create: {
      type: Boolean,
      default: false,
    },
    project: {
      required: true,
      type: Object as PropType<Project>,
    },
  },
  setup(props, { emit }) {
    const router = useRouter();
    const store = useStore();
    const state = reactive<LocalState>({
      config: {
        vcs: unknown("VCS") as VCS,
        code: "",
        token: {
          accessToken: "",
          expiresTs: 0,
          refreshToken: "",
        },
        repositoryInfo: {
          externalId: "",
          name: "",
          fullPath: "",
          webURL: "",
        },
        repositoryConfig: {
          baseDirectory: "",
          branchFilter: "",
        },
      },
      currentStep: CHOOSE_PROVIDER_STEP,
    });

    const allowNext = computed((): boolean => {
      if (state.currentStep == CONFIGURE_DEPLOY_STEP) {
        return !isEmpty(state.config.repositoryConfig.branchFilter.trim());
      }
      return true;
    });

    const tryChangeStep = (
      oldStep: number,
      newStep: number,
      allowChangeCallback: () => void
    ) => {
      state.currentStep = newStep;
      allowChangeCallback();
    };

    const tryFinishSetup = (allowFinishCallback: () => void) => {
      const createFunc = () => {
        const repositoryCreate: RepositoryCreate = {
          vcsId: state.config.vcs.id,
          projectId: props.project.id,
          name: state.config.repositoryInfo.name,
          fullPath: state.config.repositoryInfo.fullPath,
          webURL: state.config.repositoryInfo.webURL,
          baseDirectory: state.config.repositoryConfig.baseDirectory,
          branchFilter: state.config.repositoryConfig.branchFilter,
          externalId: state.config.repositoryInfo.externalId,
          accessToken: state.config.token.accessToken,
          expiresTs: state.config.token.expiresTs,
          refreshToken: state.config.token.refreshToken,
        };
        store
          .dispatch("repository/createRepository", repositoryCreate)
          .then(() => {
            allowFinishCallback();
            emit("finish");
          });
      };
      if (props.create) {
        createFunc();
      } else {
        // It's simple to implement change behavior as delete followed by create.
        // Though the delete can succeed while the create fails, this is rare, and
        // even it happens, user can still configure it again.
        store
          .dispatch("repository/deleteRepositoryByProjectId", props.project.id)
          .then(() => {
            createFunc();
          });
      }
    };

    const cancel = () => {
      emit("cancel");
      router.push({
        name: "workspace.project.detail",
        params: {
          projectSlug: projectSlug(props.project),
        },
        hash: "#version-control",
      });
    };

    return {
      state,
      stepList,
      allowNext,
      tryChangeStep,
      tryFinishSetup,
      cancel,
    };
  },
};
</script>
