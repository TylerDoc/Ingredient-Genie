# Deployments

Ansible is the tool of choice to ensure the configuration of our webserver and deploy releases.

The related code can be found in the playbooks directory

```text
playbooks/
├── inventory/
│   └── inventory.ini
├── templates/
│   ├── Caddyfile.j2
│   └── ingredient-genie.service.j2
├── configure.yml
└── deploy.yml
```

The `inventory` directory contains the current information of the webserver.

The `templates` directory contains the jinja templates for deploying the Ingredient Genie application.

The `configure.yml` playbook performs some simple hardening steps and creates a service user.

The `deploy.yml` playbook deploys Caddy to reverse proxy the application and the Ingredient Genie application itself.

To apply the configuration and deployment `cd` into the `playbooks` directory next to this one and run.

```bash
for task in configure deploy; do
    ansible-playbook -i ./inventory/inventory.ini $task.yml -u <username> --private-key=<path-to-key-file>
done
```