<script setup lang="ts">
import { useRouter } from 'vue-router';
import { $t } from '@/locales';

defineOptions({
  name: 'CreativityBanner'
});

const router = useRouter();

function goSampleData() {
  router.push({ name: 'setting_sample-data' });
}
</script>

<template>
  <div class="quick-card">
    <span class="quick-glow"></span>
    <div class="quick-icon"><SvgIcon icon="mdi:database-outline" class="text-30px" /></div>
    <h3 class="quick-title">{{ $t('page.home.quickTitle') }}</h3>
    <p class="quick-desc">{{ $t('page.home.quickDesc') }}</p>
    <div class="quick-meta">
      <div class="quick-meta-item">
        <strong>24</strong>
        <span>{{ $t('page.home.projectCount') }}</span>
      </div>
      <div class="quick-meta-item">
        <strong>6</strong>
        <span>{{ $t('page.home.message') }}</span>
      </div>
    </div>
    <NButton class="quick-btn" type="primary" block round @click="goSampleData">
      {{ $t('page.home.goSampleData') }}
      <template #icon><SvgIcon icon="mdi:arrow-right" /></template>
    </NButton>
  </div>
</template>

<style scoped>
.quick-card {
  position: relative;
  display: flex;
  flex-direction: column;
  height: 100%;
  padding: 24px;
  overflow: hidden;
  border: 1px solid rgba(100, 108, 255, 0.12);
  border-radius: 16px;
  background: linear-gradient(150deg, rgba(100, 108, 255, 0.09), rgba(139, 92, 246, 0.04));
}

.quick-glow {
  position: absolute;
  top: -70px;
  right: -70px;
  width: 190px;
  height: 190px;
  border-radius: 50%;
  background: rgba(100, 108, 255, 0.14);
  pointer-events: none;
}

.quick-icon {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 56px;
  height: 56px;
  border-radius: 16px;
  color: #fff;
  background: linear-gradient(135deg, #646cff, #8b5cf6);
  box-shadow: 0 10px 22px rgba(100, 108, 255, 0.3);
}

.quick-title {
  position: relative;
  margin: 18px 0 6px;
  font-size: 17px;
  font-weight: 600;
}

.quick-desc {
  position: relative;
  margin: 0;
  font-size: 13px;
  color: #8b91a1;
}

.quick-meta {
  position: relative;
  display: flex;
  gap: 28px;
  margin: 20px 0 22px;
}

.quick-meta-item {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.quick-meta-item strong {
  font-size: 20px;
  font-weight: 600;
}

.quick-meta-item span {
  font-size: 12px;
  color: #8b91a1;
}

.quick-btn {
  position: relative;
  margin-top: auto;
}
</style>
