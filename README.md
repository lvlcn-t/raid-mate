# raid-mate - A Discord bot for managing chores and tasks<!-- omit from toc -->

<!-- markdownlint-disable MD033 -->
<p align="center">
    <a href="/../../commits/" title="Last Commit"><img alt="Last Commit" src="https://img.shields.io/github/last-commit/lvlcn-t/raid-mate?style=flat"></a>
    <a href="/../../issues" title="Open Issues"><img alt="Open Issues" src="https://img.shields.io/github/issues/lvlcn-t/raid-mate?style=flat"></a>
    <a href="/../../pulls" title="Open Pull Requests"><img alt="Open Pull Requests" src="https://img.shields.io/github/issues-pr/lvlcn-t/raid-mate?style=flat"></a>
</p>
<!-- markdownlint-enable MD033 -->

`raid-mate` is a Discord bot that helps you manage chores and tasks in your Discord server. It is designed to be highly extensible and configurable, allowing you to customize it to your needs.

- [About this component](#about-this-component)
- [Installation](#installation)
  - [Binary](#binary)
  - [Container Image](#container-image)
  - [Helm](#helm)
- [Usage](#usage)
  - [Image](#image)
- [Configuration](#configuration)
  - [Bot Configuration](#bot-configuration)
  - [Services Configuration](#services-configuration)
  - [API Configuration](#api-configuration)
  - [Logging Configuration](#logging-configuration)
  - [Example Configuration](#example-configuration)
- [Code of Conduct](#code-of-conduct)
- [Working Language](#working-language)
- [Support and Feedback](#support-and-feedback)
- [How to Contribute](#how-to-contribute)
- [Licensing](#licensing)

## About this component

_tbd_

## Installation

The `raid-mate` bot is provided as a binary, a container image, and a Helm chart.

You can refer to the following sections for installation instructions.

**Make sure to replace `${VERSION}` with the version you want to download.**

### Binary

The binary is available for download on the [releases page](/../../releases). You can download the binary for your platform from there or use following command to install it.

```bash
curl https://github.com/lvlcn-t/raid-mate/releases/download/${VERSION}/raid-mate_${VERSION}_linux_amd64.tar.gz -Lo raid-mate.tar.gz && \
tar -xzf raid-mate.tar.gz -C ~/.local/bin && \
rm raid-mate.tar.gz
```

### Container Image

The container image is available on the [GitHub Container Registry](../../packages/). You can pull the image using the following command:

```bash
docker pull ghcr.io/lvlcn-t/raid-mate:${VERSION}
```

### Helm

You can install the `raid-mate` bot via the Helm chart. The chart is available as OCI image on the [GitHub Container Registry](../../packages/).

It can be installed using the following command:

```bash
helm upgrade -n raid-mate -i raid-mate oci://ghcr.io/lvlcn-t/charts/raid-mate:${VERSION} --create-namespace --set config.bot.token=${RAIDMATE_BOT_TOKEN}
```

Make sure to have the `RAIDMATE_BOT_TOKEN` environment variable set to the Discord bot token.

The default values are suitable for the simplest setup without any configured services. You can find all available configurations for the helm chart in the [chart's README](./chart/README.md) or the [values.yaml](./chart/values.yaml) file.

To provide the container image with secrets, you need to manually create a secret containing the environment variables. You can then utilize the `envFromSecrets` field in the `values.yaml` to enable access to the secrets. Please avoid adding sensitive information directly to the `values.yaml` file.

## Usage

To run the `raid-mate` bot, you need to provide a configuration file. The configuration file is a YAML file that contains the configuration for the bot. To learn more about the configuration options, please refer to the [configuration section](#configuration).

After you have created the configuration file, you can start the bot using the following command:

```bash
raid-mate --config /path/to/config.yaml
```

If you don't provide a configuration file, the bot will look for a file named `config.yaml` in `~/.config/raidmate/config.yaml`.

### Image

You can also run the bot using the container image. To run the bot using the container image, you can use the following command:

```bash
docker run -v /path/to/config.yaml:/config/config.yaml ghcr.io/lvlcn-t/raid-mate:${VERSION} --config /config/config.yaml
```

## Configuration

To configure the `raid-mate` bot, you need to provide a configuration file. The configuration file is a YAML file that contains the configuration for several components of the application.

Each configuration can also be provided as an environment variable with the prefix `RAIDMATE_` followed by the configuration tree. For example, the configuration `bot.token` can be provided as the environment variable `RAIDMATE_BOT_TOKEN`.

### Bot Configuration

The bot configuration is used to configure the bot itself. The following configuration options are available:

<!-- [Discord documentation](https://discord.com/developers/docs/topics/gateway#privileged-intents) -->
| Key                        | Description                                                                                                                                                                        | Type     | Default Value | Mandatory |
| -------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------- | ------------- | --------- |
| `bot.token`                | The Discord bot token                                                                                                                                                              | `string` |               | X         |
| `bot.intents.unprivileged` | Whether the bot has unprivileged intents. If not set, the bot will have no intents.                                                                                                | `bool`   | `false`       |           |
| `bot.intents.privileged`   | The list of privileged intents the bot should have. For a list of intents, see the [Discord documentation](https://discord.com/developers/docs/topics/gateway#privileged-intents). | `list`   | `[]`          |           |

### Services Configuration

To be able to use the bot, you need to configure services. The services are used to provide the bot with its external functionality. The following services are available:

- `feedback`: A service that allows users to provide feedback to the bot.

The following configuration options are available for each service:

| Key                              | Description                                                                                              | Type     | Default Value | Mandatory |
| -------------------------------- | -------------------------------------------------------------------------------------------------------- | -------- | ------------- | --------- |
| `services.feedback.service`      | Where to send the feedback to. Options: `all`, `github`, `dm`. If not set, the feedback will be ignored. | `list`   | `[]`          |           |
| `services.feedback.github.owner` | The owner of the GitHub repository where the feedback should be sent.                                    | `string` |               |           |
| `services.feedback.github.repo`  | The name of the GitHub repository where the feedback should be sent in form of an issue.                 | `string` |               |           |
| `services.feedback.dm.id`        | The Discord user ID to send the feedback to via DM. Make sure to declare it as a string.                 | `string` |               |           |

### API Configuration

The API configuration is used to configure the API that the bot should expose. If enabled you can use the API to interact with discord as well as the bot itself. The following configuration options are available:

| Key                     | Description                                                              | Type     | Default Value | Mandatory |
| ----------------------- | ------------------------------------------------------------------------ | -------- | ------------- | --------- |
| `api.enabled`           | Whether the API should be enabled. If not set, the API will be disabled. | `bool`   | `true`        |           |
| `api.address`           | The address the API should listen on.                                    | `string` | `:8080`       |           |
| `api.auth.clientId`     | The client ID for the OAuth2 client.                                     | `string` |               |           |
| `api.auth.clientSecret` | The client secret for the OAuth2 client.                                 | `string` |               |           |
| `api.auth.issuer`       | The issuer for the OAuth2 client.                                        | `string` |               |           |

### Logging Configuration

To see all the configuration options for the logging, please refer to the documentation of the [logging library](https://github.com/lvlcn-t/loggerhead?tab=readme-ov-file#configuration-via-environment-variables).

### Example Configuration

<!-- markdownlint-disable MD033 -->
<details>

<summary>Click here to see an example configuration</summary>

```yaml
# The configuration for the bot
bot:
  # The token of the discord bot
  token: ""
  intents:
    # Whether the bot has unprivileged intents
    unprivileged: true
    # The list of privileged intents the bot has
    privileged: []

# The configuration for the services
services:
  # The configuration of the feedback service
  feedback:
    # The list of services to enable the feedback service for
    # Options: all, github, dm
    # If you want to disable the feedback service set this to an empty list
    service: []
    # The configuration of the github feedback service
    github:
      # The owner of the repository to send feedback to
      owner: ""
      # The name of the repository to send feedback to
      repo: ""
      # The token to authenticate with the github api
      # Preferably use a personal access token scoped to the repository
      token: ""
    # The configuration of the direct message feedback service
    dm:
      # The id of the user to send feedback to in a direct message
      id: ""

# The configuration for the api
api:
  # Whether the api should be enabled
  enabled: true
  # The address the api should listen on
  address: :8080
  # The configuration for the authentication
  auth:
    # The client id for the oauth2 client
    clientId: ""
    # The client secret for the oauth2 client
    clientSecret: ""
    # The issuer for the oauth2 client
    issuer: ""
```

</details>
<!-- markdownlint-enable MD033 -->

## Code of Conduct

This project has adopted the [Contributor Covenant](https://www.contributor-covenant.org/) in version 2.1 as our code of
conduct. Please see the details in our [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md). All contributors must abide by the code
of conduct.

## Working Language

We decided to apply _English_ as the primary project language.

Consequently, all content will be made available primarily in English.
We also ask all interested people to use English as the preferred language to create issues,
in their code (comments, documentation, etc.) and when you send requests to us.
The application itself and all end-user facing content will be made available in other languages as needed.

## Support and Feedback

The following channels are available for discussions, feedback, and support requests:

| Type       | Channel                                                                                                                   |
| ---------- | ------------------------------------------------------------------------------------------------------------------------- |
| **Issues** | [![General Discussion](https://img.shields.io/github/issues/lvlcn-t/lvlcn-t?style=flat-square)](/../../issues/new/choose) |

## How to Contribute

Contribution and feedback is encouraged and always welcome. For more information about how to contribute, the project
structure, as well as additional contribution information, see our [Contribution Guidelines](./CONTRIBUTING.md). By
participating in this project, you agree to abide by its [Code of Conduct](./CODE_OF_CONDUCT.md) at all times.

## Licensing

Copyright (c) 2024 lvlcn-t.

Licensed under the **MIT** (the "License"); you may not use this file except in compliance with
the License.

You may obtain a copy of the License at <https://www.mit.edu/~amini/LICENSE.md>.

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "
AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the [LICENSE](./LICENSE) for
the specific language governing permissions and limitations under the License.
