<template>
    <section class="lists">
      <header class="columns page-header">
        <div class="column is-10">
          <h1 class="title is-4">
            {{ $t('globals.terms.segments') }}
            <span v-if="!isNaN(segments.total)">({{ segments.total }})</span>
          </h1>
        </div>
        <div class="column has-text-right">
          <b-field expanded>
            <b-button expanded type="is-primary" icon-left="plus" class="btn-new" @click="showNewForm" data-cy="btn-new">
              {{ $t('globals.buttons.new') }}
            </b-button>
          </b-field>
        </div>
      </header>

      <b-table :data="segments.results" :loading="loading.segments" hoverable default-sort="createdAt" paginated
        backend-pagination pagination-position="both" @page-change="onPageChange" :current-page="queryParams.page"
        :per-page="segments.perPage" :total="segments.total" backend-sorting @sort="onSort">
        <template #top-left>
          <div class="columns">
            <div class="column is-6">
              <form @submit.prevent="getSegments">
                <div>
                  <b-field>
                    <b-input v-model="queryParams.query" name="query" expanded icon="magnify" ref="query" data-cy="query" />
                    <p class="controls">
                      <b-button native-type="submit" type="is-primary" icon-left="magnify" data-cy="btn-query" />
                    </p>
                  </b-field>
                </div>
              </form>
            </div>
          </div>
        </template>

        <b-table-column v-slot="props" field="name" :label="$t('globals.fields.name')" header-class="cy-name" sortable
          width="25%" paginated backend-pagination pagination-position="both" :td-attrs="$utils.tdID"
          @page-change="onPageChange">
          <div>
            <a :href="`/segments/${props.row.id}`" @click.prevent="showEditForm(props.row)">
              {{ props.row.name }}
            </a>
          </div>
        </b-table-column>

        <b-table-column v-slot="props" field="created_at" :label="$t('globals.fields.createdAt')"
          header-class="cy-created_at" sortable>
          {{ $utils.niceDate(props.row.createdAt) }}
        </b-table-column>
        <b-table-column v-slot="props" field="updated_at" :label="$t('globals.fields.updatedAt')"
          header-class="cy-updated_at" sortable>
          {{ $utils.niceDate(props.row.updatedAt) }}
        </b-table-column>

        <b-table-column v-slot="props" cell-class="actions" align="right">
          <div>
            <a href="#" @click.prevent="showEditForm(props.row)" data-cy="btn-edit"
              :aria-label="$t('globals.buttons.edit')">
              <b-tooltip :label="$t('globals.buttons.edit')" type="is-dark">
                <b-icon icon="pencil-outline" size="is-small" />
              </b-tooltip>
            </a>

            <a href="#" @click.prevent="deleteSegment(props.row)" data-cy="btn-delete"
              :aria-label="$t('globals.buttons.delete')">
              <b-tooltip :label="$t('globals.buttons.delete')" type="is-dark">
                <b-icon icon="trash-can-outline" size="is-small" />
              </b-tooltip>
            </a>
          </div>
        </b-table-column>

        <template #empty v-if="!loading.segments">
          <empty-placeholder />
        </template>
      </b-table>

      <!-- Add / edit form modal -->
      <b-modal scroll="keep" :aria-modal="true" :active.sync="isFormVisible" :width="600" @close="onFormClose">
        <segment-form :data="curItem" :is-editing="isEditing" @finished="formFinished" />
      </b-modal>

      <p v-if="settings['app.cache_slow_queries']" class="has-text-grey">
        *{{ $t('globals.messages.slowQueriesCached') }}
        <a href="https://listmonk.app/maintenance/performance/" target="_blank" rel="noopener noreferer"
          class="has-text-grey">
          <b-icon icon="link-variant" /> {{ $t('globals.buttons.learnMore') }}
        </a>
      </p>
    </section>
</template>

<script>
import Vue from 'vue';
import { mapState } from 'vuex';
import EmptyPlaceholder from '../components/EmptyPlaceholder.vue';
import SegmentForm from './SegmentForm.vue';

export default Vue.extend({
  components: {
    SegmentForm,
    EmptyPlaceholder,
  },

  data() {
    return {
      // Current segment item being edited.
      curItem: null,
      isEditing: false,
      isFormVisible: false,
      segments: [],
      queryParams: {
        page: 1,
        query: '',
        orderBy: 'id',
        order: 'asc',
      },
    };
  },

  methods: {
    onPageChange(p) {
      this.queryParams.page = p;
      this.getSegments();
    },

    onSort(field, direction) {
      this.queryParams.orderBy = field;
      this.queryParams.order = direction;
      this.getSegments();
    },

    // Show the edit segment form.
    showEditForm(segment) {
      this.curItem = segment;
      this.isFormVisible = true;
      this.isEditing = true;
    },

    // Show the new segment form.
    showNewForm() {
      this.curItem = {};
      this.isFormVisible = true;
      this.isEditing = false;
    },

    formFinished() {
      this.getSegments();
    },

    onFormClose() {
      if (this.$route.params.id) {
        this.$router.push({ name: 'segments' });
      }
    },

    getSegments() {
      this.$api.querySegments({
        page: this.queryParams.page,
        query: this.queryParams.query.replace(/[^\p{L}\p{N}\s]/gu, ' '),
        order_by: this.queryParams.orderBy,
        order: this.queryParams.order,
      }).then((resp) => {
        this.segments = resp;
      });

      // Also fetch the minimal segments for the global store that appears
      // in dropdown menus on other pages like import and campaigns.
      this.$api.getSegments({ minimal: true, per_page: 'all' });
    },

    deleteSegment(segment) {
      this.$utils.confirm(
        this.$t('segments.confirmDelete'),
        () => {
          this.$api.deleteSegment(segment.id).then(() => {
            this.getSegments();

            this.$utils.toast(this.$t('globals.messages.deleted', { name: segment.name }));
          });
        },
      );
    },
  },

  computed: {
    ...mapState(['loading', 'settings']),
  },

  mounted() {
    if (this.$route.params.id) {
      this.$api.getSegment(parseInt(this.$route.params.id, 10)).then((data) => {
        this.showEditForm(data);
      });
    } else {
      this.getSegments();
    }
  },
});
</script>
