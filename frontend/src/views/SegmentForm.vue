<template>
  <form @submit.prevent="onSubmit">
    <div class="modal-card content" style="width: auto">
      <header class="modal-card-head">
        <p v-if="isEditing" class="has-text-grey-light is-size-7">
          {{ $t('globals.fields.id') }}: <copy-text :text="`${data.id}`" />
          {{ $t('globals.fields.uuid') }}: <copy-text :text="data.uuid" />
        </p>
        <h4 v-if="isEditing">
          {{ data.name }}
        </h4>
        <h4 v-else>
          {{ $t('segments.newSegment') }}
        </h4>
      </header>
      <section expanded class="modal-card-body">
        <b-field :label="$t('globals.fields.name')" label-position="on-border">
          <b-input :maxlength="200" :ref="'focus'" v-model="form.name" name="name"
            :placeholder="$t('globals.fields.name')" required />
        </b-field>

        <b-field :label="$t('segments.criteria')" label-position="on-border">
          <b-input :maxlength="2000" v-model="form.segmentQuery" name="segmentQuery" type="textarea"
            placeholder="subscribers.name LIKE '%user%' or subscribers.status='blocklisted'" />
          </b-field>
        <b-field :label="$t('globals.fields.description')" label-position="on-border">
          <b-input :maxlength="2000" v-model="form.description" name="description" type="textarea"
            :placeholder="$t('globals.fields.description')" />
        </b-field>
      </section>
      <footer class="modal-card-foot has-text-right">
        <b-button @click="$parent.close()">
          {{ $t('globals.buttons.close') }}
        </b-button>
        <b-button @click="countSegment()" type="is-primary" :loading="loading.segments" data-cy="btn-count">
          {{ $t('segments.buttons.count') }}
        </b-button>
        <b-button native-type="submit" type="is-primary" :loading="loading.segments" data-cy="btn-save">
          {{ $t('globals.buttons.save') }}
        </b-button>
      </footer>
    </div>
  </form>
</template>

<script>
import Vue from 'vue';
import { mapState } from 'vuex';
import CopyText from '../components/CopyText.vue';

export default Vue.extend({
  name: 'SegmentForm',

  components: {
    CopyText,
  },

  props: {
    data: { type: Object, default: () => ({}) },
    isEditing: { type: Boolean, default: false },
  },

  data() {
    return {
      // Binds form input values.
      form: {
        name: '',
        segmentQuery: '',
      },
    };
  },

  methods: {
    onSubmit() {
      if (this.isEditing) {
        this.updateSegment();
        return;
      }

      this.createSegment();
    },

    createSegment() {
      const formData = {
        name: this.form.name,
        segment_query: this.form.segmentQuery,
        description: this.form.description,
      };
      this.$api.createSegment(formData).then((data) => {
        this.$emit('finished');
        this.$parent.close();
        this.$utils.toast(this.$t('globals.messages.created', { name: data.name }));
      });
    },

    updateSegment() {
      const formData = {
        name: this.form.name,
        segment_query: this.form.segmentQuery,
        description: this.form.description,
      };
      this.$api.updateSegment({ id: this.data.id, ...formData }).then((data) => {
        this.$emit('finished');
        this.$parent.close();
        this.$utils.toast(this.$t('globals.messages.updated', { name: data.name }));
      });
    },

    countSegment() {
      this.$api.countSubscribersInSegment({ segment_query: this.form.segmentQuery }).then((data) => {
        this.$utils.confirm(`Total subscribers in this segment is [${data.total}]`)
      });
    },
  },

  computed: {
    ...mapState(['loading']),
  },

  mounted() {
    this.form = { ...this.form, ...this.$props.data };

    this.$nextTick(() => {
      this.$refs.focus.focus();
    });
  },
});
</script>
