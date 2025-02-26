<template>
  <div class="flex flex-col space-y-4">
    <div class="text-center textinfolabel">
      For database having migration history, we list up to 5 most recent
      histories below. You can click the database name to view all histories.
    </div>
    <template v-if="state.migrationHistorySectionList.length > 0">
      <MigrationHistoryTable
        :mode="'PROJECT'"
        :historySectionList="state.migrationHistorySectionList"
      />
    </template>
    <template v-else>
      <!-- This example requires Tailwind CSS v2.0+ -->
      <div class="text-center">
        <svg
          class="mx-auto w-16 h-16 text-control-light"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
          xmlns="http://www.w3.org/2000/svg"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4"
          ></path>
        </svg>
        <h3 class="mt-2 text-sm font-medium text-main">
          Do not find migration history from any database in this project.
        </h3>
        <p class="mt-1 text-sm text-control-light">
          Migration history is recorded whenever the database schema is altered.
        </p>
      </div>
    </template>
  </div>
</template>

<script lang="ts">
import { PropType, reactive, ref, watchEffect } from "@vue/runtime-core";
import { useStore } from "vuex";
import MigrationHistoryTable from "../components/MigrationHistoryTable.vue";
import {
  Database,
  InstanceMigration,
  MigrationHistory,
  Project,
} from "../types";
import { useRouter } from "vue-router";
import { BBTableSectionDataSource } from "../bbkit/types";
import { fullDatabasePath } from "../utils";

// Show at most 5 recent migration history for each database
const MAX_MIGRAION_HISTORY_COUNT = 5;

interface LocalState {
  migrationHistorySectionList: BBTableSectionDataSource<MigrationHistory>[];
}

export default {
  name: "ProjectMigrationHistoryPanel",
  components: { MigrationHistoryTable },
  props: {
    project: {
      required: true,
      type: Object as PropType<Project>,
    },
    databaseList: {
      required: true,
      type: Object as PropType<Database[]>,
    },
  },
  setup(props, ctx) {
    const searchField = ref();

    const store = useStore();
    const router = useRouter();

    const state = reactive<LocalState>({
      migrationHistorySectionList: [],
    });

    const fetchMigrationHistory = (databaseList: Database[]) => {
      state.migrationHistorySectionList = [];
      for (const database of databaseList) {
        store
          .dispatch("instance/checkMigrationSetup", database.instance.id)
          .then((migration: InstanceMigration) => {
            if (migration.status == "OK") {
              store
                .dispatch("instance/fetchMigrationHistory", {
                  instanceId: database.instance.id,
                  databaseName: database.name,
                  limit: MAX_MIGRAION_HISTORY_COUNT,
                })
                .then((migrationHistoryList: MigrationHistory[]) => {
                  if (migrationHistoryList.length > 0) {
                    const title = `${database.name} (${database.instance.environment.name})`;
                    const index = state.migrationHistorySectionList.findIndex(
                      (item: BBTableSectionDataSource<MigrationHistory>) => {
                        return item.title == title;
                      }
                    );
                    const newItem = {
                      title: title,
                      link: fullDatabasePath(database),
                      list: migrationHistoryList,
                    };
                    if (index >= 0) {
                      state.migrationHistorySectionList[index] = newItem;
                    } else {
                      state.migrationHistorySectionList.push(newItem);
                    }
                  }
                });
            }
          });
      }
    };

    const prepareMigrationHistoryList = () => {
      fetchMigrationHistory(props.databaseList);
    };
    watchEffect(prepareMigrationHistoryList);

    return {
      searchField,
      state,
    };
  },
};
</script>
