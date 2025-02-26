<template>
  <aside>
    <h2 class="sr-only">Details</h2>
    <div class="grid gap-y-6 gap-x-6 grid-cols-3">
      <template v-if="!create">
        <h2 class="textlabel flex items-center col-span-1 col-start-1">
          Status
        </h2>
        <div class="col-span-2">
          <span class="flex items-center space-x-2">
            <IssueStatusIcon :issueStatus="issue.status" :size="'normal'" />
            <span class="text-main capitalize">
              {{ issue.status.toLowerCase() }}
            </span>
          </span>
        </div>
      </template>

      <h2 class="textlabel flex items-center col-span-1 col-start-1">
        Assignee<span v-if="create" class="text-red-600">*</span>
      </h2>
      <!-- Only DBA can be assigned to the issue -->
      <div class="col-span-2">
        <MemberSelect
          :disabled="!allowEditAssignee"
          :selectedId="create ? issue.assigneeId : issue.assignee?.id"
          :allowedRoleList="['OWNER', 'DBA']"
          @select-principal-id="
            (principalId) => {
              $emit('update-assignee-id', principalId);
            }
          "
        />
      </div>

      <template v-for="(field, index) in inputFieldList" :key="index">
        <h2 class="textlabel flex items-center col-span-1 col-start-1">
          {{ field.name }}
          <span v-if="field.required" class="text-red-600">*</span>
        </h2>
        <div class="col-span-2">
          <template v-if="field.type == 'String'">
            <BBTextField
              class="text-sm"
              :disabled="!allowEditCustomField(field)"
              :required="true"
              :value="fieldValue(field)"
              :placeholder="field.placeholder"
              @end-editing="(text) => trySaveCustomField(field, text)"
            />
          </template>
          <template v-else-if="field.type == 'Boolean'">
            <BBSwitch
              :disabled="!allowEditCustomField(field)"
              :value="fieldValue(field)"
              @toggle="
                (on) => {
                  trySaveCustomField(field, on);
                }
              "
            />
          </template>
        </div>
      </template>
    </div>
    <div
      class="
        mt-6
        border-t border-block-border
        pt-6
        grid
        gap-y-6 gap-x-6
        grid-cols-3
      "
    >
      <template v-if="showStageSelect">
        <h2 class="textlabel flex items-center col-span-1 col-start-1">
          Stage
        </h2>
        <div class="col-span-2">
          <StageSelect
            :pipeline="issue.pipeline"
            :selectedId="selectedStage.id"
            @select-stage-id="(stageId) => $emit('select-stage-id', stageId)"
          />
        </div>
      </template>

      <template v-if="database">
        <h2 class="textlabel flex items-center col-span-1 col-start-1">
          Database
        </h2>
        <router-link
          :to="`/db/${databaseSlug(database)}`"
          class="col-span-2 text-sm font-medium text-main hover:underline"
        >
          {{ database.name }}
        </router-link>
      </template>

      <template v-if="showInstance">
        <h2 class="textlabel flex items-center col-span-1 col-start-1">
          Instance
        </h2>
        <router-link
          :to="`/instance/${instanceSlug(instance)}`"
          class="col-span-2 text-sm font-medium text-main hover:underline"
        >
          {{ instanceName(instance) }}
        </router-link>
      </template>

      <h2 class="textlabel flex items-center col-span-1 col-start-1">
        Environment
      </h2>
      <router-link
        :to="`/environment/${environmentSlug(environment)}`"
        class="col-span-2 text-sm font-medium text-main hover:underline"
      >
        {{ environmentName(environment) }}
      </router-link>
    </div>
    <div
      class="
        mt-6
        border-t border-block-border
        pt-6
        grid
        gap-y-6 gap-x-6
        grid-cols-3
      "
    >
      <h2 class="textlabel flex items-center col-span-1 col-start-1">
        Project
      </h2>
      <router-link
        :to="`/project/${projectSlug(project)}`"
        class="col-span-2 text-sm font-medium text-main hover:underline"
      >
        {{ projectName(project) }}
      </router-link>

      <template v-if="!create">
        <h2 class="textlabel flex items-center col-span-1 col-start-1">
          Updated
        </h2>
        <span class="textfield col-span-2">
          {{ moment(issue.updatedTs * 1000).format("LLL") }}</span
        >

        <h2 class="textlabel flex items-center col-span-1 col-start-1">
          Created
        </h2>
        <span class="textfield col-span-2">
          {{ moment(issue.createdTs * 1000).format("LLL") }}</span
        >
        <h2 class="textlabel flex items-center col-span-1 col-start-1">
          Creator
        </h2>
        <ul class="col-span-2">
          <li class="flex justify-start items-center space-x-2">
            <div class="flex-shrink-0">
              <PrincipalAvatar :principal="issue.creator" :size="'SMALL'" />
            </div>
            <router-link
              :to="`/u/${issue.creator.id}`"
              class="text-sm font-medium text-main hover:underline"
            >
              {{ issue.creator.name }}
            </router-link>
          </li>
        </ul>
      </template>
    </div>
    <IssueSubscriberPanel
      v-if="!create"
      :issue="issue"
      @add-subscriber-id="
        (subscriberId) => $emit('add-subscriber-id', subscriberId)
      "
      @remove-subscriber-id="
        (subscriberId) => $emit('remove-subscriber-id', subscriberId)
      "
    />
  </aside>
</template>

<script lang="ts">
import { computed, PropType, reactive } from "vue";
import { useStore } from "vuex";
import isEqual from "lodash-es/isEqual";
import DatabaseSelect from "../components/DatabaseSelect.vue";
import EnvironmentSelect from "../components/EnvironmentSelect.vue";
import ProjectSelect from "../components/ProjectSelect.vue";
import PrincipalAvatar from "../components/PrincipalAvatar.vue";
import MemberSelect from "../components/MemberSelect.vue";
import StageSelect from "../components/StageSelect.vue";
import IssueStatusIcon from "../components/IssueStatusIcon.vue";
import IssueSubscriberPanel from "../components/IssueSubscriberPanel.vue";
import { InputField } from "../plugins";
import {
  Database,
  Environment,
  Principal,
  Project,
  Issue,
  IssueCreate,
  EMPTY_ID,
  Stage,
  StageCreate,
  Instance,
  ONBOARDING_ISSUE_ID,
} from "../types";
import { allTaskList, isDBAOrOwner } from "../utils";

interface LocalState {}

export default {
  name: "IssueSidebar",
  emits: [
    "update-assignee-id",
    "add-subscriber-id",
    "remove-subscriber-id",
    "update-custom-field",
    "select-stage-id",
  ],
  props: {
    issue: {
      required: true,
      type: Object as PropType<Issue | IssueCreate>,
    },
    create: {
      required: true,
      type: Boolean,
    },
    selectedStage: {
      required: true,
      type: Object as PropType<Stage | StageCreate>,
    },
    inputFieldList: {
      required: true,
      type: Object as PropType<InputField[]>,
    },
    allowEdit: {
      required: true,
      type: Boolean,
    },
  },
  components: {
    DatabaseSelect,
    ProjectSelect,
    EnvironmentSelect,
    PrincipalAvatar,
    MemberSelect,
    StageSelect,
    IssueStatusIcon,
    IssueSubscriberPanel,
  },
  setup(props, { emit }) {
    const store = useStore();

    const state = reactive<LocalState>({});

    const currentUser = computed(() => store.getters["auth/currentUser"]());

    const fieldValue = (field: InputField): string => {
      return props.issue.payload[field.id];
    };

    const database = computed((): Database | undefined => {
      if (props.create) {
        const stage = props.selectedStage as StageCreate;
        if (stage.taskList[0].databaseId) {
          return store.getters["database/databaseById"](
            stage.taskList[0].databaseId
          );
        }
        return undefined;
      }
      const stage = props.selectedStage as Stage;
      return stage.taskList[0].database;
    });

    const instance = computed((): Instance => {
      if (props.create) {
        // If database is available, then we derive the instance from database because we always fetch database's instance.
        // On the other hand, instance for stage.taskList[0].instanceId might not be loaded (e.g. when creating an update schema issue)
        if (database.value) {
          return database.value.instance;
        }
        const stage = props.selectedStage as StageCreate;
        return store.getters["instance/instanceById"](
          stage.taskList[0].instanceId
        );
      }
      const stage = props.selectedStage as Stage;
      return stage.taskList[0].instance;
    });

    const environment = computed((): Environment => {
      if (props.create) {
        const stage = props.selectedStage as StageCreate;
        return store.getters["environment/environmentById"](
          stage.environmentId
        );
      }
      const stage = props.selectedStage as Stage;
      return stage.environment;
    });

    const project = computed((): Project => {
      if (props.create) {
        return store.getters["project/projectById"](
          (props.issue as IssueCreate).projectId
        );
      }
      return (props.issue as Issue).project;
    });

    const showStageSelect = computed((): boolean => {
      return (
        !props.create && allTaskList((props.issue as Issue).pipeline).length > 1
      );
    });

    const showInstance = computed((): boolean => {
      return isDBAOrOwner(currentUser.value.role);
    });

    const allowEditAssignee = computed(() => {
      // We allow the current assignee or DBA to re-assign the issue.
      // Though only DBA can be assigned to the issue, the current
      // assignee might not have DBA role in case its role is revoked after
      // being assigned to the issue.
      return (
        props.create ||
        ((props.issue as Issue).id != ONBOARDING_ISSUE_ID &&
          (props.issue as Issue).status == "OPEN" &&
          (currentUser.value.id == (props.issue as Issue).assignee?.id ||
            isDBAOrOwner(currentUser.value.role)))
      );
    });

    const allowEditCustomField = (field: InputField) => {
      return props.allowEdit && (props.create || field.allowEditAfterCreation);
    };

    const trySaveCustomField = (field: InputField, value: string | boolean) => {
      if (!isEqual(value, fieldValue(field))) {
        emit("update-custom-field", field, value);
      }
    };

    return {
      EMPTY_ID,
      state,
      fieldValue,
      environment,
      instance,
      database,
      project,
      showInstance,
      showStageSelect,
      allowEditAssignee,
      allowEditCustomField,
      trySaveCustomField,
    };
  },
};
</script>
