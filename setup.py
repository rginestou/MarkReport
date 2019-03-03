import setuptools

with open("README.md", "r") as fh:
    long_description = fh.read()

setuptools.setup(
    name="mark-report-rginestou",
    version="0.1.0",
    author="Romain Ginestou",
    author_email="romain.ginestou@gmail.com",
    description="Convert Mardown to elegant PDF reports",
    long_description=long_description,
    long_description_content_type="text/markdown",
    url="https://github.com/rginestou/MarkReport",
    packages=setuptools.find_packages(),
    classifiers=[
        "Programming Language :: Python :: 3",
        "License :: OSI Approved :: MIT License",
        "Operating System :: Linux",
    ],
)
