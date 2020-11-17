import os

import hcl2
import pytest
from mach.commands import generate

from tests.utils import get_file, get_json


@pytest.fixture
def tf_mock(mocker):
    return mocker.patch("mach.terraform.run_terraform")


def test_generate(click_runner, click_dir, tf_mock):
    result = click_runner.invoke(generate, ["-f", get_file("aws_config1.yml")])
    assert result.exit_code == 0

    deployments_dir = os.path.join(click_dir, "deployments", "aws_config1")
    sites = os.listdir(deployments_dir)
    assert sorted(sites) == ["mach-site-eu", "mach-site-us"]
    assert tf_mock.call_count == 2

    with open(os.path.join(deployments_dir, "mach-site-eu", "site.tf")) as f:
        site_config = hcl2.load(f)
    assert site_config == get_json("aws_config1_expected_mach-site-eu.json")

    with open(os.path.join(deployments_dir, "mach-site-us", "site.tf")) as f:
        site_config = hcl2.load(f)
    assert site_config == get_json("aws_config1_expected_mach-site-us.json")