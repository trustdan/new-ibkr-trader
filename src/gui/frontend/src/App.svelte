<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import SystemStatusBar from './lib/components/SystemStatusBar.svelte';
  import ScannerPanel from './lib/components/ScannerPanel.svelte';
  import { apiService } from './lib/services/api';
  import { systemHealth, isHealthy } from './lib/stores/system';

  let activeTab = 'scanner';
  let connectionTested = false;
  let testResult = '';

  onMount(() => {
    // Set the theme on mount
    document.documentElement.setAttribute('data-theme', 'ibkr-dark');
    
    // Start health monitoring when app starts
    apiService.startHealthMonitoring();
  });

  onDestroy(() => {
    // Clean up health monitoring
    apiService.stopHealthMonitoring();
  });

  async function testConnection() {
    connectionTested = true;
    try {
      const connected = await apiService.testConnection();
      testResult = connected ? 'Backend connection successful!' : 'Backend connection failed';
    } catch (error) {
      testResult = `Connection test failed: ${error.message}`;
    }
  }

  function setActiveTab(tab: string) {
    activeTab = tab;
  }

  async function mockScan() {
    try {
      const result = await apiService.scanOptions({
        symbol: 'SPY',
        filters: [{ type: 'delta', params: { min: 0.15, max: 0.35 } }],
        max_results: 50
      });
      console.log('Scan result:', result);
    } catch (error) {
      console.error('Scan failed:', error);
    }
  }
</script>

<main class="min-h-screen bg-base-100">
  <!-- System Status Bar -->
  <SystemStatusBar />

  <!-- Main Navigation -->
  <div class="navbar bg-base-200 shadow-lg border-b border-base-300">
    <div class="navbar-start">
      <div class="dropdown">
        <div tabindex="0" role="button" class="btn btn-ghost lg:hidden">
          <svg class="w-5 h-5" stroke="currentColor" fill="none" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h8m-8 6h16" />
          </svg>
        </div>
        <ul tabindex="0" class="menu menu-sm dropdown-content mt-3 z-[1] p-2 shadow bg-base-100 rounded-box w-52">
          <li><a href="#" on:click={() => setActiveTab('scanner')}>Scanner</a></li>
          <li><a href="#" on:click={() => setActiveTab('trades')}>Trades</a></li>
          <li><a href="#" on:click={() => setActiveTab('portfolio')}>Portfolio</a></li>
          <li><a href="#" on:click={() => setActiveTab('settings')}>Settings</a></li>
        </ul>
      </div>
      <a class="btn btn-ghost text-xl">
        <svg class="w-8 h-8 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6"/>
        </svg>
        IBKR Trader
      </a>
    </div>
    
    <div class="navbar-center hidden lg:flex">
      <ul class="menu menu-horizontal px-1">
        <li><a href="#" 
              class:active={activeTab === 'scanner'}
              on:click={() => setActiveTab('scanner')}>
          Scanner
        </a></li>
        <li><a href="#" 
              class:active={activeTab === 'trades'}
              on:click={() => setActiveTab('trades')}>
          Trades
        </a></li>
        <li><a href="#" 
              class:active={activeTab === 'portfolio'}
              on:click={() => setActiveTab('portfolio')}>
          Portfolio
        </a></li>
        <li><a href="#" 
              class:active={activeTab === 'settings'}
              on:click={() => setActiveTab('settings')}>
          Settings
        </a></li>
      </ul>
    </div>
    
    <div class="navbar-end">
      <div class="flex items-center gap-2">
        <div class="indicator">
          <span class="indicator-item badge badge-xs" 
                class:badge-success={$isHealthy}
                class:badge-error={!$isHealthy}>
          </span>
          <div class="btn btn-ghost btn-circle">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-5 5-5-5h5v-8h5l-5-5-5 5h5v8z"/>
            </svg>
          </div>
        </div>
      </div>
    </div>
  </div>

  <!-- Main Content Area -->
  <div class="container mx-auto p-6 max-w-7xl">
    {#if activeTab === 'scanner'}
      <ScannerPanel />
      
    {:else if activeTab === 'trades'}
      <div class="card bg-base-200 shadow-xl">
        <div class="card-body">
          <h2 class="card-title text-2xl mb-4">Trading Panel</h2>
          <div class="text-center py-12">
            <svg class="w-20 h-20 mx-auto mb-6 opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 9V7a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2m2 4h10a2 2 0 002-2v-6a2 2 0 00-2-2H9a2 2 0 00-2 2v6a2 2 0 002 2zm7-5a2 2 0 11-4 0 2 2 0 014 0z"/>
            </svg>
            <p class="text-xl mb-2">Trading panel coming soon</p>
            <p class="text-base-content/60">Execute spreads and manage orders</p>
          </div>
        </div>
      </div>

    {:else if activeTab === 'portfolio'}
      <div class="card bg-base-200 shadow-xl">
        <div class="card-body">
          <h2 class="card-title text-2xl mb-4">Portfolio Management</h2>
          <div class="text-center py-12">
            <svg class="w-20 h-20 mx-auto mb-6 opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"/>
            </svg>
            <p class="text-xl mb-2">Portfolio tracking coming soon</p>
            <p class="text-base-content/60">Monitor positions and P&L</p>
          </div>
        </div>
      </div>

    {:else if activeTab === 'settings'}
      <div class="card bg-base-200 shadow-xl">
        <div class="card-body">
          <h2 class="card-title text-2xl mb-6">Settings & Configuration</h2>
          
          <!-- Connection Test Section -->
          <div class="divider text-lg">Connection Test</div>
          <div class="form-control w-full max-w-md">
            <button 
              class="btn btn-outline btn-lg"
              class:loading={connectionTested && !testResult}
              disabled={connectionTested && !testResult}
              on:click={testConnection}
            >
              Test Backend Connection
            </button>
            
            {#if connectionTested && testResult}
              <div class="alert mt-4" 
                   class:alert-success={testResult.includes('successful')}
                   class:alert-error={!testResult.includes('successful')}>
                <span>{testResult}</span>
              </div>
            {/if}
          </div>

          <!-- System Health Section -->
          <div class="divider text-lg">System Health</div>
          <div class="stats stats-vertical lg:stats-horizontal shadow bg-base-300">
            <div class="stat">
              <div class="stat-title">TWS Connection</div>
              <div class="stat-value text-lg" 
                   class:text-success={$systemHealth.tws.connected}
                   class:text-error={!$systemHealth.tws.connected}>
                {$systemHealth.tws.connected ? 'Connected' : 'Disconnected'}
              </div>
              <div class="stat-desc">
                {$systemHealth.tws.connected ? `Uptime: ${$systemHealth.tws.uptime}s` : 'Check TWS status'}
              </div>
            </div>
            
            <div class="stat">
              <div class="stat-title">Market Data</div>
              <div class="stat-value text-lg">
                {$systemHealth.subscriptions.active}/{$systemHealth.subscriptions.max}
              </div>
              <div class="stat-desc">{$systemHealth.subscriptions.usage_pct}% usage</div>
            </div>
            
            <div class="stat">
              <div class="stat-title">Queue Status</div>
              <div class="stat-value text-lg">
                {$systemHealth.queue.size} items
              </div>
              <div class="stat-desc">
                {$systemHealth.queue.processing ? 'Processing' : 'Idle'}
              </div>
            </div>
          </div>

          <!-- Configuration Options -->
          <div class="divider text-lg">Configuration</div>
          <div class="text-center py-12">
            <svg class="w-20 h-20 mx-auto mb-6 opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"/>
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/>
            </svg>
            <p class="text-xl mb-2">Advanced settings coming soon</p>
            <p class="text-base-content/60">Configure TWS connection, risk management, and more</p>
          </div>
        </div>
      </div>
    {/if}
  </div>
</main>

<style>
  .menu a.active {
    background-color: hsl(var(--p));
    color: hsl(var(--pc));
    font-weight: 600;
  }
</style>
