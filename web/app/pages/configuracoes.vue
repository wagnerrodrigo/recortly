<template>
  <div class="max-w-2xl space-y-6">
    <div>
      <h1 class="text-2xl font-bold tracking-tight">Configurações</h1>
      <p class="mt-1 text-sm" style="color: var(--color-muted)">Conta e informações do projeto.</p>
    </div>

    <section class="card p-5">
      <h2 class="font-semibold">Conta</h2>
      <dl class="mt-3 space-y-2 text-sm">
        <div class="flex justify-between gap-4">
          <dt style="color: var(--color-muted)">Nome</dt>
          <dd class="font-medium">{{ session?.name }}</dd>
        </div>
        <div class="flex justify-between gap-4">
          <dt style="color: var(--color-muted)">E-mail</dt>
          <dd class="font-medium">{{ session?.email }}</dd>
        </div>
      </dl>
      <button class="btn btn-ghost mt-4" @click="session.logout()">Sair da conta</button>
    </section>

    <section class="card p-5">
      <h2 class="font-semibold">Regras de publicação</h2>
      <ul class="mt-3 space-y-2 text-sm" style="color: var(--color-muted)">
        <li>• Só publique conteúdo próprio ou com permissão do criador.</li>
        <li>• A licença e a origem precisam estar preenchidas na Biblioteca.</li>
        <li>• O crédito é adicionado automaticamente na descrição.</li>
        <li>• Ao remover um vídeo, os cortes e renders ligados a ele são apagados.</li>
      </ul>
    </section>

    <section class="card p-5">
      <h2 class="font-semibold">Stack</h2>
      <p class="mt-2 text-sm" style="color: var(--color-muted)">
        Nuxt 4 · Go · PostgreSQL · Redis (asynq) · FFmpeg. O plano completo está em <code class="rounded bg-zinc-100 px-1">PLANO.md</code>.
      </p>
    </section>
  </div>
</template>

<script setup lang="ts">
const session = useSession()
</script>
