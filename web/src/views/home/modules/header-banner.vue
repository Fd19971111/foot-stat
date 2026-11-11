<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import dayjs from 'dayjs';
import { fetchHello } from '@/service/api';
import { $t } from '@/locales';
import { useAuthStore } from '@/store/modules/auth';

defineOptions({
  name: 'HeaderBanner'
});

const authStore = useAuthStore();
const status = ref<'loading' | 'online' | 'offline'>('loading');
const backendDetail = ref('');

const today = dayjs().format('YYYY / MM / DD');

const statusText = computed(() => {
  if (status.value === 'loading') return $t('page.home.backendConnecting');
  return status.value === 'online' ? $t('page.home.backendOnline') : $t('page.home.backendOffline');
});

const statisticData = computed(() => [
  { id: 0, label: $t('page.home.projectCount'), value: '24' },
  { id: 1, label: $t('page.home.todo'), value: '4' },
  { id: 2, label: $t('page.home.message'), value: '6' }
]);

onMounted(async () => {
  const { data, error } = await fetchHello();
  status.value = error ? 'offline' : 'online';
  backendDetail.value = error ? '' : String(data ?? '');
});
</script>

<template>
  <div class="hero">
    <span class="hero-glow hero-glow-one"></span>
    <span class="hero-glow hero-glow-two"></span>
    <div class="hero-content">
      <div class="hero-main">
        <div class="hero-badge">
          <SvgIcon icon="mdi:chart-box-outline" class="text-14px" />
          {{ $t('page.home.overviewTitle') }}
        </div>
        <h2 class="hero-title">{{ $t('page.home.greeting', { userName: authStore.userInfo.userName }) }}</h2>
        <p class="hero-desc">{{ $t('page.home.weatherDesc') }}</p>
        <div class="hero-meta">
          <span class="status-pill" :class="status">
            <i></i>
            {{ statusText }}
            <template v-if="backendDetail">· {{ backendDetail }}</template>
          </span>
          <span class="hero-date">{{ today }}</span>
        </div>
      </div>
      <div class="hero-stats">
        <div v-for="item in statisticData" :key="item.id" class="hero-stat">
          <strong>{{ item.value }}</strong>
          <span>{{ item.label }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.hero {
  position: relative;
  overflow: hidden;
  padding: 30px 34px;
  border-radius: 18px;
  color: #fff;
  background: linear-gradient(115deg, #4f46e5 0%, #646cff 46%, #8b5cf6 100%);
  box-shadow: 0 16px 36px rgba(90, 92, 220, 0.22);
}

.hero-glow {
  position: absolute;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.16);
  pointer-events: none;
}

.hero-glow-one {
  top: -110px;
  right: 8%;
  width: 300px;
  height: 300px;
}

.hero-glow-two {
  bottom: -150px;
  left: 30%;
  width: 260px;
  height: 260px;
  background: rgba(255, 255, 255, 0.08);
}

.hero-content {
  position: relative;
  z-index: 1;
  display: flex;
  flex-wrap: wrap;
  gap: 24px;
  align-items: center;
  justify-content: space-between;
}

.hero-badge {
  display: inline-flex;
  gap: 6px;
  align-items: center;
  padding: 4px 12px;
  border: 1px solid rgba(255, 255, 255, 0.28);
  border-radius: 999px;
  font-size: 12px;
  letter-spacing: 1px;
  background: rgba(255, 255, 255, 0.12);
}

.hero-title {
  margin: 14px 0 6px;
  font-size: 26px;
  font-weight: 600;
  line-height: 1.35;
}

.hero-desc {
  margin: 0;
  font-size: 13px;
  color: rgba(255, 255, 255, 0.82);
}

.hero-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 14px;
  align-items: center;
  margin-top: 18px;
  font-size: 12px;
}

.status-pill {
  display: inline-flex;
  gap: 7px;
  align-items: center;
  padding: 5px 12px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.16);
}

.status-pill i {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #a7f3d0;
  box-shadow: 0 0 0 3px rgba(167, 243, 208, 0.25);
}

.status-pill.offline i {
  background: #fde68a;
  box-shadow: 0 0 0 3px rgba(253, 230, 138, 0.25);
}

.status-pill.loading i {
  background: rgba(255, 255, 255, 0.7);
  box-shadow: none;
}

.hero-date {
  color: rgba(255, 255, 255, 0.78);
  letter-spacing: 1px;
}

.hero-stats {
  display: flex;
  gap: 12px;
}

.hero-stat {
  display: flex;
  flex-direction: column;
  gap: 6px;
  min-width: 92px;
  padding: 14px 16px;
  border: 1px solid rgba(255, 255, 255, 0.18);
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.12);
  backdrop-filter: blur(6px);
}

.hero-stat strong {
  font-size: 24px;
  font-weight: 600;
  line-height: 1.1;
}

.hero-stat span {
  font-size: 12px;
  color: rgba(255, 255, 255, 0.8);
}

@media (max-width: 768px) {
  .hero {
    padding: 24px 20px;
  }

  .hero-title {
    font-size: 21px;
  }

  .hero-stats {
    width: 100%;
  }

  .hero-stat {
    flex: 1;
    min-width: 0;
  }
}
</style>
