<template>
  <BBTable
    :columnList="COLUMN_LIST"
    :dataSource="projectList"
    :showHeader="true"
    :leftBordered="false"
    :rightBordered="false"
    @click-row="clickProject"
  >
    <template v-slot:header>
      <BBTableHeaderCell
        class="w-4 table-cell"
        :title="state.columnList[0].title"
      />
      <BBTableHeaderCell
        class="w-24 table-cell"
        :title="state.columnList[1].title"
      />
      <BBTableHeaderCell
        class="w-8 table-cell"
        :title="state.columnList[2].title"
      />
    </template>
    <template v-slot:body="{ rowData: project }">
      <BBTableCell :leftPadding="4" class="table-cell text-gray-500">
        <span class="">{{ project.key }}</span>
      </BBTableCell>
      <BBTableCell class="truncate">
        {{ projectName(project) }}
      </BBTableCell>
      <BBTableCell class="hidden md:table-cell">
        {{ humanizeTs(project.createdTs) }}
      </BBTableCell>
    </template>
  </BBTable>
</template>

<script lang="ts">
import { PropType } from "vue";
import { useRouter } from "vue-router";
import { projectSlug } from "../utils";
import { Project } from "../types";

const COLUMN_LIST = [
  {
    title: "Key",
  },
  {
    title: "Name",
  },
  {
    title: "Created",
  },
];

export default {
  name: "ProjectTable",
  components: {},
  props: {
    projectList: {
      required: true,
      type: Object as PropType<Project[]>,
    },
  },
  setup(props, ctx) {
    const router = useRouter();

    const clickProject = function (section: number, row: number) {
      const project = props.projectList[row];
      router.push(`/project/${projectSlug(project)}`);
    };

    return {
      COLUMN_LIST,
      clickProject,
    };
  },
};
</script>
