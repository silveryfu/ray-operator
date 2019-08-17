if __name__ == '__main__':
    import sys
    from .cluster import new_spec

    if len(sys.argv) > 1:
        new_spec(sys.argv[1], create_bucket=False)
