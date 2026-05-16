<template>
  <div class="conversation-dashboard">
    <!-- 过滤栏 -->
    <div class="d-flex align-center mb-4 flex-wrap ga-2">
      <v-chip-group v-model="kindFilter" mandatory selected-class="text-primary">
        <v-chip value="" variant="outlined" size="small" class="filter-chip" filter>ALL</v-chip>
        <v-chip value="messages" variant="outlined" size="small" color="purple" class="filter-chip" filter>MESSAGES</v-chip>
        <v-chip value="chat" variant="outlined" size="small" color="blue" class="filter-chip" filter>CHAT</v-chip>
        <v-chip value="images" variant="outlined" size="small" color="pink" class="filter-chip" filter>IMAGES</v-chip>
        <v-chip value="responses" variant="outlined" size="small" color="teal" class="filter-chip" filter>RESPONSES</v-chip>
        <v-chip value="gemini" variant="outlined" size="small" color="orange" class="filter-chip" filter>GEMINI</v-chip>
      </v-chip-group>
      <v-spacer />
      <span class="text-caption text-medium-emphasis">
        Active: {{ filteredConversations.length }}
        <span v-if="overrideCount > 0" class="ml-2 text-warning">Override: {{ overrideCount }}</span>
      </span>
    </div>

    <!-- Loading -->
    <div v-if="loading && !conversations.length" class="d-flex justify-center py-12">
      <v-progress-circular indeterminate color="primary" />
    </div>

    <!-- Empty -->
    <v-card v-else-if="!filteredConversations.length" variant="outlined" class="text-center pa-12">
      <v-icon size="48" color="grey">mdi-chat-outline</v-icon>
      <div class="text-body-1 mt-4 text-medium-emphasis">
        {{ t('cockpit.empty') }}
      </div>
    </v-card>

    <!-- Conversation cards -->
    <v-row v-else>
      <v-col v-for="conv in filteredConversations" :key="conv.id" cols="12" md="6">
        <ConversationCard
          :conversation="conv"
          :override="overrides[conv.id]"
          :available-channels="getChannelsForKind(conv.kind)"
          :expanded="expandedCards.has(conv.id)"
          @toggle-expand="toggleExpand(conv.id)"
          @set-override="handleSetOverride"
          @remove-override="handleRemoveOverride"
        />
      </v-col>
    </v-row>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { api, type ConversationInfo, type SequenceOverrideInfo, type ChannelSequenceEntry } from '@/services/api'
import { useGlobalTick } from '@/composables/useGlobalTick'
import { useI18n } from '@/i18n'
import ConversationCard from './ConversationCard.vue'

const { t } = useI18n()

const loading = ref(true)
const conversations = ref<ConversationInfo[]>([])
const overrides = ref<Record<string, SequenceOverrideInfo>>({})
const kindFilter = ref('')
const expandedCards = ref(new Set<string>())
type DashboardChannel = { index: number; name: string; priority: number; status: string }

const channelsByKind = ref<Record<string, DashboardChannel[]>>({})

function normalizeChannel(ch: any): DashboardChannel {
  const index = ch.index ?? ch.Index ?? 0
  return {
    index,
    name: ch.name ?? ch.Name ?? `Channel ${index}`,
    priority: ch.priority ?? ch.Priority ?? index,
    status: ch.status ?? ch.Status ?? 'active',
  }
}

function normalizeChannelsByKind(value: Record<string, any[]>): Record<string, DashboardChannel[]> {
  return Object.fromEntries(
    Object.entries(value).map(([kind, channels]) => [
      kind,
      (channels || [])
        .map(normalizeChannel)
        .sort((a, b) => (a.priority - b.priority) || (a.index - b.index)),
    ]),
  )
}

const filteredConversations = computed(() => {
  const filter = kindFilter.value
  const items = filter ? conversations.value.filter(c => c.kind === filter) : conversations.value
  return [...items].sort((a, b) => new Date(b.lastActiveAt).getTime() - new Date(a.lastActiveAt).getTime())
})

const overrideCount = computed(() => Object.keys(overrides.value).length)

function getChannelsForKind(kind: string): DashboardChannel[] {
  return channelsByKind.value[kind] || []
}

async function fetchAllChannels() {
  const kinds = ['messages', 'chat', 'responses', 'gemini', 'images'] as const
  for (const kind of kinds) {
    try {
      const dashboard = await api.getChannelDashboard(kind)
      if (!channelsByKind.value[kind]?.length) {
        channelsByKind.value[kind] = (dashboard.channels || [])
          .map(normalizeChannel)
          .sort((a, b) => (a.priority - b.priority) || (a.index - b.index))
      }
    } catch (e) {
      console.error(`[ConversationDashboard] fetch ${kind} channels error:`, e)
    }
  }
}

async function fetchConversations() {
  try {
    const resp = await api.getConversations(kindFilter.value || undefined)
    conversations.value = resp.conversations || []
    overrides.value = resp.overrides || {}
    if (resp.channelsByKind) {
      channelsByKind.value = normalizeChannelsByKind(resp.channelsByKind)
    }
  } catch (e) {
    console.error('[ConversationDashboard] fetch error:', e)
  } finally {
    loading.value = false
  }
}

function toggleExpand(id: string) {
  const next = new Set(expandedCards.value)
  if (next.has(id)) {
    next.delete(id)
  } else {
    next.add(id)
  }
  expandedCards.value = next
}

async function handleSetOverride(convId: string, sequence: ChannelSequenceEntry[]) {
  try {
    await api.setConversationOverride(convId, sequence)
    await fetchConversations()
  } catch (e) {
    console.error('[ConversationDashboard] set override error:', e)
  }
}

async function handleRemoveOverride(convId: string) {
  try {
    await api.removeConversationOverride(convId)
    await fetchConversations()
  } catch (e) {
    console.error('[ConversationDashboard] remove override error:', e)
  }
}

// Polling
const tick = useGlobalTick(3000, 'ConversationDashboard')
tick.onTick(() => fetchConversations())
fetchConversations()
fetchAllChannels()
</script>

<style scoped>
.conversation-dashboard {
  max-width: 1400px;
  margin: 0 auto;
}
.filter-chip {
  border-radius: 0 !important;
  font-size: 10px !important;
  font-weight: 700;
  letter-spacing: 0.06em;
}
</style>