<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';

import { apolloClient } from '@/graphql/client';
import {
  ADD_OVERRIDE,
  INVITE_MEMBER,
  REMOVE_MEMBER,
  ROTATE_API_KEY,
  TEAM_QUERY,
  UPDATE_SCHEDULE,
} from '@/graphql/operations';
import { useAuthStore } from '@/stores/auth';
import type { Role, Team } from '@/types';

const authStore = useAuthStore();

const team = ref<Team | null>(null);
const loading = ref(false);
const error = ref<string | null>(null);
const toast = ref<string | null>(null);

const inviteEmail = ref('');
const inviteRole = ref<Exclude<Role, 'OWNER'>>('RESPONDER');

const rotationIds = ref<string[]>([]);
const intervalDays = ref(7);
const addToRotationId = ref('');

const overrideUserId = ref('');
const overrideStart = ref('');
const overrideEnd = ref('');
const overrideReason = ref('');

const showRotateConfirm = ref(false);
const revealedKey = ref<string | null>(null);
const keyCopied = ref(false);

const teamId = computed(() => authStore.user?.teamId ?? '');
const members = computed(() => team.value?.members ?? []);
const memberName = (id: string) => members.value.find((member) => member.id === id)?.name ?? id;
const rotationCandidates = computed(() =>
  members.value.filter((member) => !rotationIds.value.includes(member.id)),
);
const ownerCount = computed(() => members.value.filter((member) => member.role === 'OWNER').length);

function flash(message: string) {
  toast.value = message;
  window.setTimeout(() => {
    toast.value = null;
  }, 3000);
}

function reportError(err: unknown, fallback: string) {
  error.value = err instanceof Error ? err.message : fallback;
}

async function loadTeam() {
  if (!teamId.value) {
    return;
  }
  loading.value = true;
  error.value = null;
  try {
    const { data } = await apolloClient.query({
      query: TEAM_QUERY,
      variables: { id: teamId.value },
      fetchPolicy: 'network-only',
    });
    team.value = data.team;
    rotationIds.value = team.value?.onCallSchedule?.rotation.map((user) => user.id) ?? [];
    intervalDays.value = team.value?.onCallSchedule?.intervalDays ?? 7;
  } catch (err) {
    reportError(err, 'Unable to load team');
  } finally {
    loading.value = false;
  }
}

async function inviteMember() {
  if (!inviteEmail.value.trim()) {
    return;
  }
  try {
    await apolloClient.mutate({
      mutation: INVITE_MEMBER,
      variables: { teamId: teamId.value, email: inviteEmail.value.trim(), role: inviteRole.value },
    });
    inviteEmail.value = '';
    flash('Member invited');
    await loadTeam();
  } catch (err) {
    reportError(err, 'Invite failed');
  }
}

async function removeMember(userId: string) {
  if (!window.confirm('Remove this member from the team?')) {
    return;
  }
  try {
    await apolloClient.mutate({
      mutation: REMOVE_MEMBER,
      variables: { teamId: teamId.value, userId },
    });
    flash('Member removed');
    await loadTeam();
  } catch (err) {
    reportError(err, 'Remove failed');
  }
}

function moveInRotation(index: number, delta: number) {
  const target = index + delta;
  if (target < 0 || target >= rotationIds.value.length) {
    return;
  }
  const next = [...rotationIds.value];
  [next[index], next[target]] = [next[target], next[index]];
  rotationIds.value = next;
}

function removeFromRotation(id: string) {
  rotationIds.value = rotationIds.value.filter((rotationId) => rotationId !== id);
}

function addToRotation() {
  if (addToRotationId.value && !rotationIds.value.includes(addToRotationId.value)) {
    rotationIds.value = [...rotationIds.value, addToRotationId.value];
    addToRotationId.value = '';
  }
}

async function saveSchedule() {
  if (rotationIds.value.length === 0) {
    error.value = 'Add at least one member to the rotation';
    return;
  }
  try {
    await apolloClient.mutate({
      mutation: UPDATE_SCHEDULE,
      variables: { teamId: teamId.value, rotation: rotationIds.value, intervalDays: intervalDays.value },
    });
    flash('On-call schedule saved');
    await loadTeam();
  } catch (err) {
    reportError(err, 'Could not save schedule');
  }
}

async function addOverride() {
  if (!overrideUserId.value || !overrideStart.value || !overrideEnd.value) {
    error.value = 'Override requires a user, start, and end';
    return;
  }
  try {
    await apolloClient.mutate({
      mutation: ADD_OVERRIDE,
      variables: {
        teamId: teamId.value,
        userId: overrideUserId.value,
        startsAt: new Date(overrideStart.value).toISOString(),
        endsAt: new Date(overrideEnd.value).toISOString(),
        reason: overrideReason.value,
      },
    });
    overrideUserId.value = '';
    overrideStart.value = '';
    overrideEnd.value = '';
    overrideReason.value = '';
    flash('Override added');
    await loadTeam();
  } catch (err) {
    reportError(err, 'Could not add override');
  }
}

async function confirmRotateApiKey() {
  try {
    const { data } = await apolloClient.mutate({
      mutation: ROTATE_API_KEY,
      variables: { teamId: teamId.value },
    });
    showRotateConfirm.value = false;
    revealedKey.value = data.rotateApiKey;
    keyCopied.value = false;
    await loadTeam();
  } catch (err) {
    showRotateConfirm.value = false;
    reportError(err, 'Could not rotate API key');
  }
}

async function copyKey() {
  if (!revealedKey.value) {
    return;
  }
  await navigator.clipboard.writeText(revealedKey.value);
  keyCopied.value = true;
}

onMounted(loadTeam);
</script>

<template>
  <main class="min-h-screen bg-zinc-950 px-6 py-6 text-zinc-50">
    <section class="mx-auto grid max-w-4xl gap-6">
      <div class="flex items-center justify-between">
        <RouterLink
          class="text-sm text-cyan-300 hover:text-cyan-200"
          to="/dashboard"
        >
          Back to dashboard
        </RouterLink>
        <span class="text-sm text-zinc-500">{{ team?.name }}</span>
      </div>
      <h1 class="text-2xl font-semibold">
        Team settings
      </h1>

      <p
        v-if="loading"
        class="text-sm text-zinc-400"
      >
        Loading team...
      </p>
      <p
        v-if="error"
        class="rounded-md border border-red-500/40 bg-red-500/10 p-3 text-sm text-red-200"
      >
        {{ error }}
      </p>

      <!-- Members -->
      <section class="rounded-md border border-zinc-800 p-5">
        <h2 class="text-lg font-semibold">
          Members
        </h2>
        <table class="mt-4 w-full text-left text-sm">
          <thead class="text-xs uppercase text-zinc-500">
            <tr>
              <th class="py-2">
                Name
              </th>
              <th class="py-2">
                Email
              </th>
              <th class="py-2">
                Role
              </th>
              <th class="py-2">
                Joined
              </th>
              <th class="py-2" />
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="member in members"
              :key="member.id"
              class="border-t border-zinc-800"
            >
              <td class="py-2">
                {{ member.name }}
              </td>
              <td class="py-2 text-zinc-400">
                {{ member.email }}
              </td>
              <td class="py-2">
                {{ member.role }}
              </td>
              <td class="py-2 text-zinc-400">
                {{ new Date(member.createdAt).toLocaleDateString() }}
              </td>
              <td class="py-2 text-right">
                <button
                  class="rounded-md border border-red-500/40 px-2 py-1 text-xs text-red-200 disabled:opacity-30"
                  type="button"
                  :disabled="member.id === authStore.user?.id || (member.role === 'OWNER' && ownerCount <= 1)"
                  @click="removeMember(member.id)"
                >
                  Remove
                </button>
              </td>
            </tr>
          </tbody>
        </table>

        <form
          class="mt-4 flex flex-wrap items-end gap-3"
          @submit.prevent="inviteMember"
        >
          <div class="grow">
            <label
              class="text-xs text-zinc-500"
              for="invite-email"
            >Invite by email</label>
            <input
              id="invite-email"
              v-model="inviteEmail"
              class="mt-1 w-full rounded-md border border-zinc-800 bg-zinc-950 p-2 text-sm"
              placeholder="teammate@example.com"
              type="email"
            >
          </div>
          <select
            v-model="inviteRole"
            class="rounded-md border border-zinc-800 bg-zinc-950 p-2 text-sm"
          >
            <option value="RESPONDER">
              Responder
            </option>
            <option value="VIEWER">
              Viewer
            </option>
          </select>
          <button
            class="rounded-md bg-cyan-400 px-3 py-2 text-sm font-semibold text-zinc-950"
            type="submit"
          >
            Invite
          </button>
        </form>
      </section>

      <!-- On-call schedule -->
      <section class="rounded-md border border-zinc-800 p-5">
        <h2 class="text-lg font-semibold">
          On-call schedule
        </h2>
        <p
          v-if="team?.onCallSchedule?.currentOnCall"
          class="mt-1 text-sm text-zinc-400"
        >
          Currently on call: <span class="text-cyan-300">{{ team.onCallSchedule.currentOnCall.name }}</span>
        </p>

        <ol class="mt-4 grid gap-2">
          <li
            v-for="(id, index) in rotationIds"
            :key="id"
            class="flex items-center justify-between rounded-md bg-zinc-900 px-3 py-2 text-sm"
          >
            <span>{{ index + 1 }}. {{ memberName(id) }}</span>
            <span class="flex gap-1">
              <button
                class="rounded border border-zinc-700 px-2 text-xs disabled:opacity-30"
                type="button"
                :disabled="index === 0"
                @click="moveInRotation(index, -1)"
              >
                ↑
              </button>
              <button
                class="rounded border border-zinc-700 px-2 text-xs disabled:opacity-30"
                type="button"
                :disabled="index === rotationIds.length - 1"
                @click="moveInRotation(index, 1)"
              >
                ↓
              </button>
              <button
                class="rounded border border-red-500/40 px-2 text-xs text-red-200"
                type="button"
                @click="removeFromRotation(id)"
              >
                ✕
              </button>
            </span>
          </li>
          <li
            v-if="rotationIds.length === 0"
            class="text-sm text-zinc-500"
          >
            No rotation configured yet.
          </li>
        </ol>

        <div class="mt-3 flex flex-wrap items-end gap-3">
          <select
            v-model="addToRotationId"
            class="rounded-md border border-zinc-800 bg-zinc-950 p-2 text-sm"
            @change="addToRotation"
          >
            <option value="">
              Add member to rotation…
            </option>
            <option
              v-for="candidate in rotationCandidates"
              :key="candidate.id"
              :value="candidate.id"
            >
              {{ candidate.name }}
            </option>
          </select>
          <div>
            <label
              class="text-xs text-zinc-500"
              for="interval"
            >Rotation interval</label>
            <select
              id="interval"
              v-model.number="intervalDays"
              class="mt-1 block rounded-md border border-zinc-800 bg-zinc-950 p-2 text-sm"
            >
              <option :value="1">
                Daily
              </option>
              <option :value="7">
                Weekly
              </option>
              <option :value="14">
                Biweekly
              </option>
            </select>
          </div>
          <button
            class="rounded-md bg-cyan-400 px-3 py-2 text-sm font-semibold text-zinc-950"
            type="button"
            @click="saveSchedule"
          >
            Save schedule
          </button>
        </div>

        <!-- Overrides -->
        <h3 class="mt-6 text-sm font-semibold">
          Overrides
        </h3>
        <ul class="mt-2 grid gap-1 text-sm text-zinc-400">
          <li
            v-for="override in team?.onCallSchedule?.overrides ?? []"
            :key="override.id"
          >
            {{ override.user.name }}: {{ new Date(override.startsAt).toLocaleString() }} →
            {{ new Date(override.endsAt).toLocaleString() }}
            <span v-if="override.reason">— {{ override.reason }}</span>
          </li>
        </ul>
        <div class="mt-3 flex flex-wrap items-end gap-3">
          <select
            v-model="overrideUserId"
            class="rounded-md border border-zinc-800 bg-zinc-950 p-2 text-sm"
          >
            <option value="">
              Select user…
            </option>
            <option
              v-for="member in members"
              :key="member.id"
              :value="member.id"
            >
              {{ member.name }}
            </option>
          </select>
          <input
            v-model="overrideStart"
            class="rounded-md border border-zinc-800 bg-zinc-950 p-2 text-sm"
            type="datetime-local"
          >
          <input
            v-model="overrideEnd"
            class="rounded-md border border-zinc-800 bg-zinc-950 p-2 text-sm"
            type="datetime-local"
          >
          <input
            v-model="overrideReason"
            class="grow rounded-md border border-zinc-800 bg-zinc-950 p-2 text-sm"
            placeholder="Reason (optional)"
          >
          <button
            class="rounded-md border border-zinc-700 px-3 py-2 text-sm"
            type="button"
            @click="addOverride"
          >
            Add override
          </button>
        </div>
      </section>

      <!-- API key -->
      <section class="rounded-md border border-zinc-800 p-5">
        <h2 class="text-lg font-semibold">
          API key
        </h2>
        <p class="mt-2 font-mono text-sm text-zinc-300">
          •••• •••• •••• {{ team?.apiKeyHint || '????' }}
        </p>
        <button
          class="mt-4 rounded-md border border-amber-400/50 px-3 py-2 text-sm font-semibold text-amber-200"
          type="button"
          @click="showRotateConfirm = true"
        >
          Rotate API key
        </button>
      </section>
    </section>

    <!-- Rotate confirmation -->
    <div
      v-if="showRotateConfirm"
      class="fixed inset-0 z-10 flex items-center justify-center bg-black/60 px-4"
    >
      <div class="w-full max-w-md rounded-md border border-zinc-800 bg-zinc-950 p-6">
        <h3 class="text-lg font-semibold">
          Rotate API key?
        </h3>
        <p class="mt-2 text-sm text-zinc-400">
          The current key stops working <strong>immediately</strong>. Any webhook integrations using
          it will start returning 401 until you update them with the new key.
        </p>
        <div class="mt-5 flex justify-end gap-2">
          <button
            class="rounded-md border border-zinc-700 px-3 py-2 text-sm"
            type="button"
            @click="showRotateConfirm = false"
          >
            Cancel
          </button>
          <button
            class="rounded-md bg-amber-300 px-3 py-2 text-sm font-semibold text-zinc-950"
            type="button"
            @click="confirmRotateApiKey"
          >
            Rotate now
          </button>
        </div>
      </div>
    </div>

    <!-- One-time key reveal -->
    <div
      v-if="revealedKey"
      class="fixed inset-0 z-10 flex items-center justify-center bg-black/60 px-4"
    >
      <div class="w-full max-w-lg rounded-md border border-zinc-800 bg-zinc-950 p-6">
        <h3 class="text-lg font-semibold">
          New API key
        </h3>
        <p class="mt-2 text-sm text-red-200">
          Copy it now — this key will never be shown again.
        </p>
        <code class="mt-4 block break-all rounded-md bg-zinc-900 p-3 text-sm text-cyan-300">{{ revealedKey }}</code>
        <div class="mt-5 flex justify-end gap-2">
          <button
            class="rounded-md border border-zinc-700 px-3 py-2 text-sm"
            type="button"
            @click="copyKey"
          >
            {{ keyCopied ? 'Copied!' : 'Copy' }}
          </button>
          <button
            class="rounded-md bg-cyan-400 px-3 py-2 text-sm font-semibold text-zinc-950"
            type="button"
            @click="revealedKey = null"
          >
            Done
          </button>
        </div>
      </div>
    </div>

    <p
      v-if="toast"
      class="fixed bottom-6 right-6 rounded-md bg-emerald-400 px-4 py-2 text-sm font-semibold text-zinc-950"
    >
      {{ toast }}
    </p>
  </main>
</template>
