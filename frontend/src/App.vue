<template>
  <main class="app-shell">
    <aside class="panel patient-panel">
      <header class="panel-header">
        <div>
          <p class="eyebrow">Personal Doctor</p>
          <h1>病人</h1>
        </div>
        <button class="icon-button" title="刷新" @click="loadPatients">
          <RefreshCw :size="18" />
        </button>
      </header>

      <form class="patient-form" @submit.prevent="addPatient">
        <input v-model="patientForm.name" placeholder="姓名" required />
        <div class="split">
          <select v-model="patientForm.gender">
            <option value="">性别</option>
            <option value="男">男</option>
            <option value="女">女</option>
            <option value="其他">其他</option>
          </select>
          <input v-model="patientForm.birthday" type="date" />
        </div>
        <input v-model="patientForm.phone" placeholder="联系方式" />
        <textarea v-model="patientForm.allergies" rows="2" placeholder="过敏史"></textarea>
        <button class="primary" type="submit">
          <UserPlus :size="17" />
          新增病人
        </button>
      </form>

      <div class="patient-list">
        <button
          v-for="patient in patients"
          :key="patient.id"
          class="patient-item"
          :class="{ active: patient.id === selectedPatientId }"
          @click="selectPatient(patient.id)"
        >
          <span class="avatar">{{ patient.name.slice(0, 1) }}</span>
          <span>
            <strong>{{ patient.name }}</strong>
            <small>{{ patient.gender || '未填性别' }} · {{ patient.birthday || '未填生日' }}</small>
          </span>
        </button>
      </div>
    </aside>

    <section class="workspace" v-if="selectedPatient">
      <section class="panel records-panel">
        <header class="panel-header">
          <div>
            <p class="eyebrow">{{ selectedPatient.name }}</p>
            <h2>病历和药方</h2>
          </div>
          <span class="count">{{ records.length }} 条</span>
        </header>

        <form class="record-form" @submit.prevent="addRecord">
          <div class="split">
            <select v-model="recordForm.kind">
              <option value="condition">病情</option>
              <option value="prescription">药方</option>
              <option value="exam">检查</option>
              <option value="note">备注</option>
            </select>
            <input v-model="recordForm.title" placeholder="标题" required />
          </div>
          <textarea v-model="recordForm.content" rows="5" placeholder="录入症状、诊断、用药、检查结果等" required></textarea>
          <button class="primary" type="submit">
            <FilePlus2 :size="17" />
            保存记录
          </button>
        </form>

        <div class="record-list">
          <article v-for="record in records" :key="record.id" class="record-card">
            <div class="record-meta">
              <span>{{ kindLabel(record.kind) }}</span>
              <time>{{ formatDate(record.recordedAt) }}</time>
            </div>
            <h3>{{ record.title }}</h3>
            <p>{{ record.content }}</p>
          </article>
          <p v-if="records.length === 0" class="empty">还没有病历，先录入一条病情或药方。</p>
        </div>
      </section>

      <section class="panel chat-panel">
        <header class="panel-header">
          <div>
            <p class="eyebrow">Doctor Agent</p>
            <h2>问诊聊天</h2>
          </div>
          <Stethoscope :size="22" />
        </header>

        <div ref="messageBox" class="messages">
          <div
            v-for="message in messages"
            :key="message.id"
            class="message"
            :class="message.role"
          >
            <span>{{ message.role === 'user' ? '你' : '医生助手' }}</span>
            <p>{{ message.content }}</p>
          </div>
          <p v-if="messages.length === 0" class="empty">选择病人后可以开始问诊，助手会读取已录入病历。</p>
        </div>

        <form class="chat-input" @submit.prevent="submitMessage">
          <textarea
            v-model="draft"
            rows="3"
            placeholder="描述症状、询问药物注意事项或让助手整理病情摘要"
            @keydown.enter.exact.prevent="submitMessage"
          ></textarea>
          <button class="send-button" type="submit" :disabled="sending || !draft.trim()">
            <SendHorizontal :size="18" />
            发送
          </button>
        </form>
      </section>
    </section>

    <section class="blank" v-else>
      <Stethoscope :size="44" />
      <h2>先创建或选择一位病人</h2>
    </section>

    <p v-if="error" class="toast">{{ error }}</p>
  </main>
</template>

<script setup>
import { computed, nextTick, onMounted, reactive, ref } from 'vue'
import {
  FilePlus2,
  RefreshCw,
  SendHorizontal,
  Stethoscope,
  UserPlus,
} from 'lucide-vue-next'
import {
  createPatient,
  createRecord,
  listMessages,
  listPatients,
  listRecords,
  sendMessage,
} from './api'

const patients = ref([])
const records = ref([])
const messages = ref([])
const selectedPatientId = ref('')
const draft = ref('')
const error = ref('')
const sending = ref(false)
const messageBox = ref(null)

const patientForm = reactive({
  name: '',
  gender: '',
  birthday: '',
  phone: '',
  allergies: '',
})

const recordForm = reactive({
  kind: 'condition',
  title: '',
  content: '',
})

const selectedPatient = computed(() =>
  safeArray(patients.value).find((patient) => patient.id === selectedPatientId.value),
)

onMounted(loadPatients)

async function loadPatients() {
  await guard(async () => {
    patients.value = safeArray(await listPatients())
    if (!patients.value.some((patient) => patient.id === selectedPatientId.value)) {
      selectedPatientId.value = ''
      records.value = []
      messages.value = []
    }
    if (!selectedPatientId.value && patients.value.length > 0) {
      selectedPatientId.value = patients.value[0].id
      await loadPatientData()
    }
  })
}

async function selectPatient(id) {
  selectedPatientId.value = id
  await loadPatientData()
}

async function loadPatientData() {
  if (!selectedPatientId.value) return
  await guard(async () => {
    const [nextRecords, nextMessages] = await Promise.all([
      listRecords(selectedPatientId.value),
      listMessages(selectedPatientId.value),
    ])
    records.value = safeArray(nextRecords)
    messages.value = safeArray(nextMessages)
    await scrollMessages()
  })
}

async function addPatient() {
  await guard(async () => {
    const patient = await createPatient({ ...patientForm })
    patients.value = [patient, ...safeArray(patients.value)]
    selectedPatientId.value = patient.id
    records.value = []
    messages.value = []
    Object.assign(patientForm, { name: '', gender: '', birthday: '', phone: '', allergies: '' })
  })
}

async function addRecord() {
  if (!selectedPatientId.value) return
  await guard(async () => {
    const record = await createRecord(selectedPatientId.value, { ...recordForm })
    records.value = [record, ...safeArray(records.value)]
    Object.assign(recordForm, { kind: 'condition', title: '', content: '' })
  })
}

async function submitMessage() {
  if (!selectedPatientId.value || !draft.value.trim() || sending.value) return
  const content = draft.value.trim()
  draft.value = ''
  sending.value = true
  try {
    await guard(async () => {
      const response = await sendMessage(selectedPatientId.value, content)
      messages.value = [...safeArray(messages.value), response.user, response.assistant].filter(Boolean)
      await scrollMessages()
    })
  } finally {
    sending.value = false
  }
}

async function guard(action) {
  error.value = ''
  try {
    await action()
  } catch (err) {
    error.value = err.message || '操作失败'
  }
}

async function scrollMessages() {
  await nextTick()
  if (messageBox.value) {
    messageBox.value.scrollTop = messageBox.value.scrollHeight
  }
}

function kindLabel(kind) {
  return {
    condition: '病情',
    prescription: '药方',
    exam: '检查',
    note: '备注',
  }[kind] || '记录'
}

function formatDate(value) {
  return new Intl.DateTimeFormat('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  }).format(new Date(value))
}

function safeArray(value) {
  return Array.isArray(value) ? value : []
}
</script>
