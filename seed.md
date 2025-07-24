# Seed

## Template Selection

Choose the template from the gallery that best fits the project requirements and avaliable env variables.

for example, if you have a STRIPE_SECRET_KEY, and the project requirements include payments, you should choose the template that includes stripe.

you can choose templates that do not support supabase or vercel even if the project requirements needs them and you have them in the db, but the template needs to support next.js/nuxt.js and postgres (respectively the vercel and supabase providers). in this case, you will still need to create the issues related to these providers integration.

this logic should apply to all providers, not just the ones in the example.

prefer the more minimal templates, with less dependencies and features that what we actually need: for example, if we need a static frontend/landing page, we should not use a template that includes a database nor payments, etc.

if a project parts needs data processing, ml, specialy algorithms and models, you should prefer python for that part. 

if the description indicates that a part of the project depends on a library that is only available in specific language, you should prefer that language for that part. for example: cli project that needs pytorch or docling would be better in python.

### Gallery

Docs Sites:

https://github.com/shuding/nextra-docs-template (docs site)

Full stack or websites:

https://github.com/boxyhq/saas-starter-kit (nextjs saas app)

https://github.com/dzlau/stripe-supabase-saas-template (nextjs saas app with stripe and supabase)

https://github.com/vercel/ai-chatbot (ai chatbot with next.js, ai sdk, shadcn/ui, Neon Serverless db, Auth.js)

https://github.com/vercel/commerce (nextjs, shopify)

https://github.com/vercel/platforms (A production-ready example of a multi-tenant application built with Next.js 15, featuring custom subdomains for each tenant - nextjs, react, upstash redis, shadcn/ui)

https://github.com/nextjs/saas-starter (nextjs, stripe, Drizzle, postgres)

https://github.com/vercel/vercel/tree/main/examples/nextjs (nextjs, shadcn/ui - good for static pages, landing pages, public websites, etc)

https://github.com/vercel/vercel/tree/main/examples/nuxtjs (nuxtjs, shadcn/ui - good for static pages, landing pages, public websites, etc)


https://github.com/brocoders/nestjs-boilerplate (NestJS boilerplate. Auth, TypeORM, Mongoose, Postgres, MongoDB, Mailing, I18N)

https://github.com/async-labs/saas (SaaS boilerplate. Productive stack: React, Material-UI, Next, MobX, WebSockets, Express, Node, Mongoose, MongoDB. Written with TypeScript.)

https://github.com/kriasoft/react-starter-kit - (cloudflare workers, react, typescript, tailwind, shadcn/ui)

https://github.com/vercel/next-forge (react typescript stripe sentry nextjs neon seo feature-flags dark-mode prisma tailwindcss posthog clerk)


https://github.com/kriasoft/graphql-starter-kit - (Monorepo, GraphQL API, PostgreSQL, React, and Joy UI.)

https://github.com/lxieyang/chrome-extension-boilerplate-react (chrome extension boilerplate, react, webpack)


there are many more templates in these catalogs: https://github.com/vercel/vercel/tree/main/examples , https://github.com/topics/boilerplate , https://github.com/topics/template , https://github.com/topics/starter-kit , https://github.com/topics/starter-template , https://github.com/prisma/prisma-examples (for prisma), 

find the one with the best fit and the most stars.

for games, depending on the game engine type, you can use the following templates:

https://github.com/Avalin/Unity-CI-Templates (Unity)

https://github.com/epayet/phaser-typescript-seed (Phaser - simpler projects, 2d, typescript)

for mobile apps, you can use the following templates (you may need to use more than one template to cover both mobile and web):

https://github.com/thecodingmachine/react-native-boilerplate

https://github.com/obytes/react-native-template-obytes (React Native, Expo, PNPM, TypeScript, TailwindCSS, Husky, EAS, GitHub Actions, Env Vars, expo-router, react-query, react-hook-form)



pure microservices:

https://github.com/hagopj13/node-express-boilerplate (nodejs, express, mongoose)
https://github.com/santoshshinde2012/node-boilerplate 

packages:

https://github.com/tomchen/example-typescript-package
https://github.com/lambda-science/modern-python-boilerplate

if none of the templates fit the project requirements, try https://github.com/pankod/superplate ( generic scaffolding for many many types of stacks), and if that doesn't work, create a issue to ask for the specific templates to use or to build a new one. 

CLI/Command line tools:

https://github.com/chamoda/cookiecutter-typer (python, typer, cookiecutter template)

https://github.com/simonw/click-app (python, click, cookiecutter template)

https://github.com/samhuk/node-cli-template (Typescript, node, commander.js) 
----

cookiecutter templates (python and go):

https://github.com/search?q=cookiecutter&type=Repositories (the entire catalog of cookiecutter templates)

https://github.com/audreyfeldroy/cookiecutter-pypackage (python package, cookiecutter template)

use cookiecutter and the right cookiecutter template if the projects needs python or go. (you will need to install it first using pip install)

prefer the template by stack first, not functionality. (and never use php)

## Notes

### GitHub Workflows

The workflows are already configured to use standard npm script commands to run the tests, build, deploy, etc.
No need to change them nor copy them from the template.

## Additional GitHub Issues to open

In addition to the usual instructions, open the following issues

### Populate the sh scripts

Populate the sh scripts (./scripts directory) with the right commands according to the project stack, requirements, etc.

### Vercel project creation and association (if the stack includes vercel)

use the VERCEL_TOKEN and VERCEL_ORG_ID to generate a new project in vercel and save the VERCEL_PROJECT_ID in the repo:

as an env variable for the actions (in github)
as a comment to this ticket.
in a .vercel.env file in the repo (with only the new VERCEL_PROJECT_ID)

if the input env variables are not provided, create a issue to ask for them instead of creating the project.

then, make sure you update the npm run deploy command to:
link the project to vercel:
          vercel link --yes --token=${{ secrets.VERCEL_TOKEN }} \
          --scope=${{ vars.VERCEL_ORG_ID }} \
          --project=${{ vars.VERCEL_PROJECT_ID }}
and deploy the project:
          vercel deploy --prod --token=${{ secrets.VERCEL_TOKEN }} \
          --scope=${{ vars.VERCEL_ORG_ID }} 

### Supabase project creation and association (if the stack includes supabase)

use the SUPABASE_ACCESS_TOKEN and SUPABASE_ORG_ID to generate a new project (with default db password from SUPABASE_DB_PASSWORD) in supabase and save the SUPABASE_PROJECT_REF and SUPABASE_PROJECT_URL and SUPABASE_ANON_KEY in the repo:

as an env variable for the actions (in github)
as a comment to this ticket.
in a .supabase.env file in the repo (with only the new SUPABASE_PROJECT_REF and SUPABASE_PROJECT_URL)

also create a connection string for the database in the project
with password populated at build time from the SUPABASE_DB_PASSWORD secret.

make sure to update the code/deployment to use the connection string (DATABASE_URL) and database configuration (env variables in the target deployment service - for example vercel)

make sure the integrate everything with the actual deployment. for example, in next with vercel, you need to set the NEXT_PUBLIC_SUPABASE_ANON_KEY as env variables in the vercel project.

if the input env variables/vars/secrets (access token, org id, db password) are not provided, create a issue to ask for them instead of creating the project.

### Auth providers integration github issues to open

if the stack includes auth providers, open the following issues:

- github auth (use AUTH_GITHUB_CLIENT_ID and AUTH_GITHUB_CLIENT_SECRET from build time secrets / envs) - specifically github isn't a valid prefix for secrets. so we need to use a different prefix. all the other auth providers are valid prefixes.
- google auth (use GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET from build time secrets / envs)
- facebook auth (use FACEBOOK_CLIENT_ID and FACEBOOK_CLIENT_SECRET from build time secrets / envs)

if using an authentication provider, make sure to include integration instructions with the code base itself, for example: include instructions to set vercel or deployment env variables in the target deployment service (GITHUB_CLIENT_ID)

as well as setting service or callback urls, for example in supabase and next: NEXT_PUBLIC_SUPABASE_URL
and probably (NEXT_PUBLIC_SITE_URL)

### More issues for providers integration similar to the ones above (if the stack includes more providers)

For Example: Prisma, need to set the DATABASE_URL according to the rest of the stack during the build/run time. as well as to add prisma commands to the npm scripts. (build should could prisma generate too)

### Project specifications
(do not write the specs.md file yourself)
Create a new issue to 'Create project specifications' with the following details:

- the project specifications should be a detailed description of the project, with the following details 
  - the project name
  - seed(s) used (if any)
  - the project description
  - the project requirements
  - the project stack
  - the project deployment providers
  - instruction to research and create the project specifications in the specs.md file in the root of the repo.

### Husky pre-commit hooks (if the seed/stack includes a package.json,  otherwise, the right tool for pre-commit hooks)

1. install husky (package.json and package-lock.json)
2. configure it to run the tests and build before committing in .husky/pre-commit (according to the project npm scripts)
3. modify package.json to use the new husky commands (using 'prepare').

### Seed Specific instructions

Examine the instructions in the README.md file or the docs site of the seed/starter kit to complete the scaffolding. pay attention to details like:

1. 3rd party services and deployment services (supabase, vercel, etc) integrations setup
2. environment variables setup
3. changes in configuration files to adapt to the seed/starter kit
4. changes to the scripts dir (which are currently called by the main github workfow) to adapt to the seed/starter kit.
5. only open tickets for non-functional (like the ones above) gaps from the seed/starter kit, not product features, gaps and requirements.
6. another issue to open should be to 'Create project specifications'
the actual issue should be opened with the details above in the description. with instructions to perform the changes in the repo. (for the producer-agent)

- for each feature that is not trivial, add an issue for the developer agent to create an "implementation guide" by researching how it is implemented in similar open source projects with similar relevant stack components
- Implementation guides for remaining non-trivial features.
  
- last issue you open should be: 'Test and fix the project' , with the body:"Check that the project builds, tests and able to package for deployment. also, that it runs and serves (dev env and test env), if not,fix it. if you were unable to fix it. open a new issue and link it to this ticket, then open a pull request for what you were able to fix. if you reached a point where there are some screenshots or videos of tests (for example, playwright or cypress), attach them to this issue in a comment.' - for the developer agent to handle.
