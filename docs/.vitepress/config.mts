import { defineConfig } from 'vitepress'

export default defineConfig({
  title: 'EnvGuard',
  description: 'Validate .env files against a declarative YAML schema. Catch misconfigurations before deployment.',
  base: '/envguard/',
  lastUpdated: true,

  head: [
    ['link', { rel: 'icon', href: '/envguard/favicon.ico' }],
    ['meta', { name: 'theme-color', content: '#10b981' }],
    ['meta', { property: 'og:type', content: 'website' }],
    ['meta', { property: 'og:locale', content: 'en' }],
    ['meta', { property: 'og:title', content: 'EnvGuard | Environment Variable Validator' }],
    ['meta', { property: 'og:description', content: 'Validate .env files against a declarative YAML schema.' }],
  ],

  themeConfig: {
    logo: '/logo.svg',

    nav: [
      { text: 'Guide', link: '/guide/quickstart' },
      { text: 'Reference', link: '/reference/cli' },
      { text: 'v2.0.0', items: [
        { text: 'Changelog', link: '/reference/changelog' },
        { text: 'Contributing', link: '/guide/contributing' },
      ]},
    ],

    sidebar: {
      '/guide/': [
        {
          text: 'Getting Started',
          collapsed: false,
          items: [
            { text: 'Installation', link: '/guide/installation' },
            { text: 'Quick Start', link: '/guide/quickstart' },
            { text: 'Configuration', link: '/guide/configuration' },
          ],
        },
        {
          text: 'Core Concepts',
          collapsed: false,
          items: [
            { text: 'Schema', link: '/guide/schema' },
            { text: 'Validation Rules', link: '/guide/validation-rules' },
            { text: 'Type Coercion', link: '/guide/type-coercion' },
            { text: 'Secrets Scanning', link: '/guide/secrets' },
            { text: 'Watch Mode', link: '/guide/watch-mode' },
            { text: 'Schema Inference', link: '/guide/schema-inference' },
          ],
        },
        {
          text: 'Advanced',
          collapsed: false,
          items: [
            { text: 'Monorepo Support', link: '/guide/monorepo' },
            { text: 'Schema Composition', link: '/guide/schema-composition' },
            { text: 'Interpolation', link: '/guide/interpolation' },
            { text: 'LSP / VS Code', link: '/guide/lsp' },
            { text: 'Git Hooks', link: '/guide/git-hooks' },
          ],
        },
        {
          text: 'Integration',
          collapsed: false,
          items: [
            { text: 'CI / CD', link: '/guide/ci-cd' },
            { text: 'GitHub Action', link: '/guide/github-action' },
            { text: 'Pre-commit Hook', link: '/guide/pre-commit' },
            { text: 'Node.js Wrapper', link: '/guide/nodejs-wrapper' },
            { text: 'Python Wrapper', link: '/guide/python-wrapper' },
          ],
        },
      ],
      '/reference/': [
        {
          text: 'Reference',
          collapsed: false,
          items: [
            { text: 'CLI Commands', link: '/reference/cli' },
            { text: 'Schema Format', link: '/reference/schema-format' },
            { text: 'Exit Codes', link: '/reference/exit-codes' },
            { text: 'Config File', link: '/reference/config-file' },
            { text: 'Changelog', link: '/reference/changelog' },
          ],
        },
      ],
    },

    editLink: {
      pattern: 'https://github.com/firasmosbehi/envguard/edit/main/docs/:path',
      text: 'Edit this page on GitHub',
    },

    socialLinks: [
      { icon: 'github', link: 'https://github.com/firasmosbehi/envguard' },
    ],

    footer: {
      message: 'Released under the MIT License.',
      copyright: 'Copyright © 2024-present Firas Mosbehi',
    },

    search: {
      provider: 'local',
    },

    outline: {
      level: 'deep',
    },
  },
})
