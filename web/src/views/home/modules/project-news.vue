<script setup lang="ts">
import { computed } from 'vue';
import { $t } from '@/locales';

defineOptions({
  name: 'ProjectNews'
});

interface NewsItem {
  id: number;
  content: string;
  time: string;
  icon: string;
}

const newses = computed<NewsItem[]>(() => [
  { id: 1, content: $t('page.home.projectNews.desc1'), time: '2026-11-10 09:24', icon: 'mdi:check-circle-outline' },
  { id: 2, content: $t('page.home.projectNews.desc2'), time: '2026-11-10 08:12', icon: 'mdi:shield-check-outline' },
  { id: 3, content: $t('page.home.projectNews.desc3'), time: '2026-11-09 16:03', icon: 'mdi:database-plus-outline' },
  { id: 4, content: $t('page.home.projectNews.desc4'), time: '2026-11-09 14:30', icon: 'mdi:loading' },
  { id: 5, content: $t('page.home.projectNews.desc5'), time: '2026-11-08 20:11', icon: 'mdi:file-document-outline' }
]);

function showMore() {
  window.$message?.info($t('page.home.projectNews.moreNews'));
}
</script>

<template>
  <NCard :bordered="false" size="small" class="h-full news-card">
    <template #header>
      <div class="news-header">
        <span class="title-bar"></span>
        {{ $t('page.home.projectNews.title') }}
      </div>
    </template>
    <template #header-extra>
      <NButton quaternary size="small" type="primary" @click="showMore">
        {{ $t('page.home.projectNews.moreNews') }}
      </NButton>
    </template>
    <div class="timeline">
      <div v-for="item in newses" :key="item.id" class="timeline-item">
        <div class="timeline-dot"><SvgIcon :icon="item.icon" class="text-13px" /></div>
        <div class="timeline-body">
          <p class="timeline-text">{{ item.content }}</p>
          <span class="timeline-time">{{ item.time }}</span>
        </div>
      </div>
    </div>
  </NCard>
</template>

<style scoped>
.news-card {
  border-radius: 16px;
}

.news-header {
  display: flex;
  gap: 9px;
  align-items: center;
  font-size: 15px;
  font-weight: 600;
}

.title-bar {
  width: 4px;
  height: 15px;
  border-radius: 4px;
  background: linear-gradient(180deg, #646cff, #8b5cf6);
}

.timeline {
  position: relative;
  padding-left: 4px;
}

.timeline-item {
  position: relative;
  display: flex;
  gap: 12px;
  padding-bottom: 16px;
}

.timeline-item:last-child {
  padding-bottom: 0;
}

.timeline-item:not(:last-child)::before {
  position: absolute;
  top: 26px;
  bottom: 0;
  left: 12px;
  width: 1px;
  background: rgba(130, 140, 170, 0.22);
  content: '';
}

.timeline-dot {
  z-index: 1;
  display: flex;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  width: 25px;
  height: 25px;
  border-radius: 50%;
  color: #646cff;
  background: rgba(100, 108, 255, 0.1);
}

.timeline-body {
  flex: 1;
  padding-top: 2px;
}

.timeline-text {
  margin: 0;
  font-size: 13px;
  line-height: 1.5;
}

.timeline-time {
  font-size: 11px;
  color: #a3a8b5;
}
</style>
